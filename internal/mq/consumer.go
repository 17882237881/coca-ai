package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer Kafka 消费者
type Consumer struct {
	reader   *kafka.Reader
	handlers []MessageHandler
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, event *MessageEvent) error

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers []string // Kafka Broker 地址列表
	GroupID string   // 消费者组 ID
}

// NewConsumer 创建 Kafka 消费者
func NewConsumer(cfg *ConsumerConfig) *Consumer {
	if cfg.GroupID == "" {
		cfg.GroupID = "coca-chat-consumer"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          TopicChatMessages,
		GroupID:        cfg.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,      // 自动提交偏移量
		StartOffset:    kafka.LastOffset, // 从最新消息开始
	})

	return &Consumer{
		reader:   reader,
		handlers: make([]MessageHandler, 0),
	}
}

// RegisterHandler 注册消息处理函数
func (c *Consumer) RegisterHandler(handler MessageHandler) {
	c.handlers = append(c.handlers, handler)
}

// Start 启动消费者 (阻塞式)
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("[Kafka Consumer] Starting consumer for topic: %s", TopicChatMessages)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Kafka Consumer] Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			// 读取消息
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Printf("[Kafka Consumer] Fetch message error: %v", err)
				continue
			}

			// 解析消息
			var event MessageEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Kafka Consumer] Unmarshal error: %v, value: %s", err, string(msg.Value))
				// 提交偏移量，跳过无法解析的消息
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			// 调用所有处理函数
			for _, handler := range c.handlers {
				if err := handler(ctx, &event); err != nil {
					log.Printf("[Kafka Consumer] Handler error: %v", err)
					// 根据业务需求决定是否重试或跳过
				}
			}

			// 提交偏移量
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
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
		return c.reader.Close()
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
