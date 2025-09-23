package mq

import (
	"fmt"
	"sync"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// KafkaClientType 定义 Kafka 客户端类型
type KafkaClientType string

const (
	// GameScore 游戏分数上报
	GameScore KafkaClientType = "game_score"
	// Chat 聊天消息
	Chat KafkaClientType = "chat"
	// UserBehavior 用户行为日志
	UserBehavior KafkaClientType = "user_behavior"
	// Notification 系统通知
	Notification KafkaClientType = "notification"
)

// KafkaFactory Kafka工厂，管理所有Kafka客户端
type KafkaFactory struct {
	config *config.KafkaConfigs
	// 生产者映射表
	producers map[KafkaClientType]*KafkaProducer
	// 消费者映射表
	consumers map[KafkaClientType]*KafkaConsumer
	mu        sync.RWMutex
}

// NewKafkaFactory 创建Kafka工厂
func NewKafkaFactory(cfg *config.KafkaConfigs) *KafkaFactory {
	return &KafkaFactory{
		config:    cfg,
		producers: make(map[KafkaClientType]*KafkaProducer),
		consumers: make(map[KafkaClientType]*KafkaConsumer),
	}
}

// GetProducer 获取指定类型的生产者
func (f *KafkaFactory) GetProducer(clientType KafkaClientType) (*KafkaProducer, error) {
	f.mu.RLock()
	if producer, exists := f.producers[clientType]; exists {
		f.mu.RUnlock()
		return producer, nil
	}
	f.mu.RUnlock()

	// 如果不存在，创建新的生产者
	f.mu.Lock()
	defer f.mu.Unlock()

	// 双重检查
	if producer, exists := f.producers[clientType]; exists {
		return producer, nil
	}

	var kafkaConfig config.KafkaConfig
	switch clientType {
	case GameScore:
		kafkaConfig = f.config.GameScore
	case Chat:
		kafkaConfig = f.config.Chat
	case UserBehavior:
		kafkaConfig = f.config.UserBehavior
	case Notification:
		kafkaConfig = f.config.Notification
	default:
		return nil, fmt.Errorf("unknown kafka client type: %s", clientType)
	}

	producer, err := NewKafkaProducer(KafkaConfig{
		Brokers: kafkaConfig.Brokers,
		Topic:   kafkaConfig.Topic,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer for %s: %w", clientType, err)
	}

	f.producers[clientType] = producer
	return producer, nil
}

// GetConsumer 获取指定类型的消费者
func (f *KafkaFactory) GetConsumer(clientType KafkaClientType) (*KafkaConsumer, error) {
	f.mu.RLock()
	if consumer, exists := f.consumers[clientType]; exists {
		f.mu.RUnlock()
		return consumer, nil
	}
	f.mu.RUnlock()

	// 如果不存在，创建新的消费者
	f.mu.Lock()
	defer f.mu.Unlock()

	// 双重检查
	if consumer, exists := f.consumers[clientType]; exists {
		return consumer, nil
	}

	var kafkaConfig config.KafkaConfig
	switch clientType {
	case GameScore:
		kafkaConfig = f.config.GameScore
	case Chat:
		kafkaConfig = f.config.Chat
	case UserBehavior:
		kafkaConfig = f.config.UserBehavior
	case Notification:
		kafkaConfig = f.config.Notification
	default:
		return nil, fmt.Errorf("unknown kafka client type: %s", clientType)
	}

	consumer, err := NewKafkaConsumer(KafkaConfig{
		Brokers: kafkaConfig.Brokers,
		Topic:   kafkaConfig.Topic,
		GroupID: kafkaConfig.GroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer for %s: %w", clientType, err)
	}

	f.consumers[clientType] = consumer
	return consumer, nil
}

// Close 关闭所有客户端
func (f *KafkaFactory) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 关闭所有生产者
	for _, producer := range f.producers {
		if err := producer.Close(); err != nil {
			utils.Error("Error closing producer", zap.Error(err))
		}
	}

	// 关闭所有消费者
	for _, consumer := range f.consumers {
		if err := consumer.Close(); err != nil {
			utils.Error("Error closing consumer", zap.Error(err))
		}
	}
}
