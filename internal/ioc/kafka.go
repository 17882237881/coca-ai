package ioc

import (
	"coca-ai/internal/config"
	"coca-ai/internal/mq"
)

// InitKafkaProducer 初始化 Kafka 生产者
func InitKafkaProducer() *mq.Producer {
	cfg := config.Get()

	if len(cfg.Kafka.Brokers) == 0 {
		// 开发环境可能没有 Kafka，返回 nil
		return nil
	}

	return mq.NewProducer(&mq.ProducerConfig{
		Brokers:      cfg.Kafka.Brokers,
		RequiredAcks: cfg.Kafka.Producer.RequiredAcks,
		Async:        cfg.Kafka.Producer.Async,
		BatchSize:    cfg.Kafka.Producer.BatchSize,
		BatchTimeout: cfg.Kafka.Producer.BatchTimeoutMS,
		Compression:  cfg.Kafka.Producer.Compression,
		WriteTimeout: cfg.Kafka.Producer.WriteTimeoutMS,
		MaxAttempts:  cfg.Kafka.Producer.MaxAttempts,
	})
}

// BaseConsumer 包装器，避免 Wire 依赖冲突
type BaseConsumer struct {
	*mq.Consumer
}

// InitKafkaConsumer 初始化 Kafka 消费者
func InitKafkaConsumer() BaseConsumer {
	cfg := config.Get()

	if len(cfg.Kafka.Brokers) == 0 {
		return BaseConsumer{nil}
	}

	c := mq.NewConsumer(&mq.ConsumerConfig{
		Brokers:        cfg.Kafka.Brokers,
		GroupID:        cfg.Kafka.Consumer.GroupID,
		MinBytes:       cfg.Kafka.Consumer.MinBytes,
		MaxBytes:       cfg.Kafka.Consumer.MaxBytes,
		MaxWait:        cfg.Kafka.Consumer.MaxWaitMS,
		StartOffset:    cfg.Kafka.Consumer.StartOffset,
		MaxRetry:       cfg.Kafka.Consumer.MaxRetry,
		RetryBackoffMS: cfg.Kafka.Consumer.RetryBackoffMS,
		CommitTimeout:  cfg.Kafka.Consumer.CommitTimeoutMS,
		DLQTopic:       cfg.Kafka.DLQTopic,
	})
	return BaseConsumer{Consumer: c}
}

// BindKafkaHandlers 绑定 Kafka 消费处理器
func BindKafkaHandlers(c BaseConsumer, handler *mq.MessagePersistHandler) *mq.Consumer {
	consumer := c.Consumer
	if consumer == nil {
		return nil
	}
	if handler != nil {
		consumer.RegisterHandler(handler.Handle)
	}
	return consumer
}
