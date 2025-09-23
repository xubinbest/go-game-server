package mq

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// KafkaProducer Kafka生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(cfg KafkaConfig) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

// SendMessage 发送消息
func (p *KafkaProducer) SendMessage(ctx context.Context, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	return nil
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(cfg KafkaConfig) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumer: group,
		topic:    cfg.Topic,
	}, nil
}

// Message Kafka消息
type Message struct {
	Key   []byte
	Value []byte
}

// ConsumeMessages 消费消息
func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, handler func(msg Message) error) error {
	topics := []string{c.topic}
	for {
		err := c.consumer.Consume(ctx, topics, &consumerGroupHandler{handler: handler})
		if err != nil {
			return fmt.Errorf("consume error: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	return nil
}

// consumerGroupHandler 实现 sarama.ConsumerGroupHandler 接口
type consumerGroupHandler struct {
	handler func(msg Message) error
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.handler(Message{
			Key:   msg.Key,
			Value: msg.Value,
		})
		if err != nil {
			// 记录错误但继续处理
			utils.Error("Error processing message", zap.Error(err))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
