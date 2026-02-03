package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	// TopicChatMessages 聊天消息 Topic
	TopicChatMessages = "chat.messages"
	// TopicChatMessagesDLQ 聊天消息死信 Topic
	TopicChatMessagesDLQ = "chat.messages.dlq"
)

// MessageEvent Kafka 消息事件结构
type MessageEvent struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"` // Unix 毫秒
}

// Producer Kafka 生产者
type Producer struct {
	writer *kafka.Writer
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Brokers      []string // Kafka Broker 地址列表
	RequiredAcks string
	Async        bool
	BatchSize    int
	BatchTimeout int // 毫秒
	Compression  string
	WriteTimeout int // 毫秒
	MaxAttempts  int
}

// NewProducer 创建 Kafka 生产者
func NewProducer(cfg *ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        TopicChatMessages,
		Balancer:     &kafka.Hash{}, // 分区器，Hash 策略
		BatchSize:    orDefaultInt(cfg.BatchSize, 100), // 批量大小
		BatchTimeout: time.Duration(orDefaultInt(cfg.BatchTimeout, 10)) * time.Millisecond, // 批量超时时间
		RequiredAcks: parseRequiredAcks(cfg.RequiredAcks), 
		Async:        cfg.Async, // 是否异步
		Compression:  parseCompression(cfg.Compression), // 压缩方式
		WriteTimeout: time.Duration(orDefaultInt(cfg.WriteTimeout, 10000)) * time.Millisecond, // 写入超时时间
		MaxAttempts:  orDefaultInt(cfg.MaxAttempts, 10), // 最大重试次数
	}

	return &Producer{writer: writer}
}

// SendMessage 发送消息事件到 Kafka
func (p *Producer) SendMessage(ctx context.Context, event *MessageEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", event.SessionID)), // 按 SessionID 分区
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	return nil
}

// SendMessages 批量发送消息事件
func (p *Producer) SendMessages(ctx context.Context, events []*MessageEvent) error {
	messages := make([]kafka.Message, len(events))
	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event failed: %w", err)
		}
		messages[i] = kafka.Message{
			Key:   []byte(fmt.Sprintf("%d", event.SessionID)),
			Value: data,
		}
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("write messages failed: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

func parseRequiredAcks(value string) kafka.RequiredAcks {
	switch value {
	case "all", "ALL", "require_all":
		return kafka.RequireAll
	case "none", "NONE", "require_none":
		return kafka.RequireNone
	case "one", "ONE", "require_one":
		return kafka.RequireOne
	default:
		return kafka.RequireAll
	}
}

func parseCompression(value string) kafka.Compression {
	switch value {
	case "gzip", "GZIP":
		return kafka.Gzip
	case "lz4", "LZ4":
		return kafka.Lz4
	case "zstd", "ZSTD":
		return kafka.Zstd
	case "none", "NONE":
		return kafka.Compression(0)
	default:
		return kafka.Snappy
	}
}
