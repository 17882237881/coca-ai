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
		Brokers: cfg.Kafka.Brokers,
	})
}

// InitKafkaConsumer 初始化 Kafka 消费者
func InitKafkaConsumer() *mq.Consumer {
	cfg := config.Get()

	if len(cfg.Kafka.Brokers) == 0 {
		// 开发环境可能没有 Kafka，返回 nil
		return nil
	}

	return mq.NewConsumer(&mq.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers,
		GroupID: "coca-chat-consumer",
	})
}
