package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer Kafka 消费者
type Consumer struct {
	reader        *kafka.Reader
	handlers      []MessageHandler
	dlqWriter     *kafka.Writer
	maxRetry      int
	retryBackoff  time.Duration
	commitTimeout time.Duration
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, event *MessageEvent) error

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers        []string // Kafka Broker 地址列表
	GroupID        string   // 消费者组 ID
	MinBytes       int
	MaxBytes       int
	MaxWait        int // 毫秒
	StartOffset    string
	MaxRetry       int
	RetryBackoffMS int
	CommitTimeout  int // 毫秒
	DLQTopic       string
}

// NewConsumer 创建 Kafka 消费者
func NewConsumer(cfg *ConsumerConfig) *Consumer {
	if cfg.GroupID == "" {
		cfg.GroupID = "coca-chat-consumer"
	}

	startOffset := kafka.LastOffset
	switch strings.ToLower(cfg.StartOffset) {
	case "earliest", "first":
		startOffset = kafka.FirstOffset
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          TopicChatMessages,
		GroupID:        cfg.GroupID,
		MinBytes:       orDefaultInt(cfg.MinBytes, 10e3),  // 10KB
		MaxBytes:       orDefaultInt(cfg.MaxBytes, 10e6),  // 10MB
		MaxWait:        time.Duration(orDefaultInt(cfg.MaxWait, 500)) * time.Millisecond,
		CommitInterval: 0, // 手动提交 offset
		StartOffset:    startOffset,
	})

	dlqTopic := cfg.DLQTopic
	if dlqTopic == "" {
		dlqTopic = TopicChatMessagesDLQ
	}
	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        dlqTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		AllowAutoTopicCreation: true,
	}

	return &Consumer{
		reader:        reader,
		handlers:      make([]MessageHandler, 0),
		dlqWriter:     dlqWriter,
		maxRetry:      orDefaultInt(cfg.MaxRetry, 3),
		retryBackoff:  time.Duration(orDefaultInt(cfg.RetryBackoffMS, 200)) * time.Millisecond,
		commitTimeout: time.Duration(orDefaultInt(cfg.CommitTimeout, 3000)) * time.Millisecond,
	}
}

// RegisterHandler 注册消息处理函数
func (c *Consumer) RegisterHandler(handler MessageHandler) {
	c.handlers = append(c.handlers, handler)
}

// Start 启动消费者 (阻塞式)
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("[Kafka Consumer] Starting consumer for topic: %s", TopicChatMessages)
	if len(c.handlers) == 0 {
		log.Printf("[Kafka Consumer] No handlers registered, consumer will still read and commit offsets.")
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[Kafka Consumer] Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			// 读取消息 (手动提交 Offset)
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Printf("[Kafka Consumer] Read message error: %v", err)
				continue
			}

			// 解析消息
			var event MessageEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Kafka Consumer] Unmarshal error: %v, value: %s", err, string(msg.Value))
				c.sendToDLQ(ctx, msg, err)
				c.commitMessage(ctx, msg)
				continue
			}

			if err := c.handleWithRetry(ctx, &event); err != nil {
				log.Printf("[Kafka Consumer] Handler error: %v", err)
				c.sendToDLQ(ctx, msg, err)
				c.commitMessage(ctx, msg)
				continue
			}

			if err := c.commitMessage(ctx, msg); err != nil {
				log.Printf("[Kafka Consumer] Commit error: %v", err)
			}
		}
	}
}

// StartAsync 异步启动消费者 (非阻塞)
func (c *Consumer) StartAsync(ctx context.Context) {
	go func() {
		if err := c.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[Kafka Consumer] Consumer stopped with error: %v", err)
		}
	}()
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	if c.reader != nil {
		_ = c.reader.Close()
	}
	if c.dlqWriter != nil {
		return c.dlqWriter.Close()
	}
	return nil
}

// Stats 获取消费者统计信息
func (c *Consumer) Stats() kafka.ReaderStats {
	return c.reader.Stats()
}

// String 返回消费者描述
func (c *Consumer) String() string {
	stats := c.Stats()
	return fmt.Sprintf("[Kafka Consumer] Topic: %s, Messages: %d, Errors: %d",
		TopicChatMessages, stats.Messages, stats.Errors)
}

func (c *Consumer) handleWithRetry(ctx context.Context, event *MessageEvent) error {
	if len(c.handlers) == 0 {
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetry; attempt++ {
		for _, handler := range c.handlers {
			if err := handler(ctx, event); err != nil {
				lastErr = err
				break
			}
			lastErr = nil
		}
		if lastErr == nil {
			return nil
		}
		if attempt < c.maxRetry {
			if !sleepWithContext(ctx, c.retryBackoff) {
				return ctx.Err()
			}
		}
	}

	return lastErr
}

func (c *Consumer) commitMessage(ctx context.Context, msg kafka.Message) error {
	commitCtx := ctx
	if c.commitTimeout > 0 {
		var cancel context.CancelFunc
		commitCtx, cancel = context.WithTimeout(ctx, c.commitTimeout)
		defer cancel()
	}
	return c.reader.CommitMessages(commitCtx, msg)
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, cause error) {
	if c.dlqWriter == nil {
		return
	}
	msgCopy := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: append(msg.Headers, kafka.Header{Key: "dlq_error", Value: []byte(cause.Error())}),
		Time:    time.Now(),
	}
	if err := c.dlqWriter.WriteMessages(ctx, msgCopy); err != nil {
		log.Printf("[Kafka Consumer] DLQ write failed: %v", err)
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
