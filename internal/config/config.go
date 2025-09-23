package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LoadBalancer LoadBalancerConfig `yaml:"loadBalancer"`
	GRPC         GRPCConfig         `yaml:"grpc"`
	Redis        RedisConfig        `yaml:"redis"`
	Database     DatabaseConfig     `yaml:"database"`
	Auth         AuthConfig         `yaml:"auth"`
	RateLimit    RateLimitConfig    `yaml:"rateLimit"`
	Kafka        KafkaConfig        `yaml:"kafka"`
	KafkaConfigs KafkaConfigs       `yaml:"kafka_configs"`
}

type LoadBalancerConfig struct {
	Strategy string `yaml:"strategy"` // random, leastconn, roundrobin
}

type GRPCConfig struct {
	MaxRetry       int           `yaml:"maxRetry"`
	Timeout        time.Duration `yaml:"timeout"`
	KeepAlive      time.Duration `yaml:"keepAlive"`
	MaxRecvMsgSize int           `yaml:"maxRecvMsgSize"` // in MB
	MaxSendMsgSize int           `yaml:"maxSendMsgSize"` // in MB
}

type RedisConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	PoolSize int           `yaml:"poolSize"`
	Timeout  time.Duration `yaml:"timeout"`
	Cluster  []string      `yaml:"cluster"`
}

type DatabaseConfig struct {
	MySQL   MySQLConfig   `yaml:"mysql"`
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

type MySQLConfig struct {
	Enabled         bool          `yaml:"enabled"`
	DSN             string        `yaml:"dsn"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	LogLevel        string        `yaml:"logLevel"`    // GORM日志级别: silent, error, warn, info
	AutoMigrate     bool          `yaml:"autoMigrate"` // 是否自动迁移表结构
}

type MongoDBConfig struct {
	Enabled  bool          `yaml:"enabled"`
	URI      string        `yaml:"uri"`
	Database string        `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout"`
}

type AuthConfig struct {
	SecretKey       string        `yaml:"secretKey"`
	TokenExpire     time.Duration `yaml:"tokenExpire"`
	Salt            string        `yaml:"salt"`
	TokenExpireTime time.Duration `yaml:"tokenExpireTime"`
	WhitelistPaths  []string      `yaml:"whitelistPaths"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `yaml:"requestsPerSecond"`
	Burst             int `yaml:"burst"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type KafkaConfigs struct {
	// 游戏分数上报配置
	GameScore KafkaConfig `yaml:"game_score"`
	// 聊天消息配置（示例）
	Chat KafkaConfig `yaml:"chat"`
	// 用户行为日志配置（示例）
	UserBehavior KafkaConfig `yaml:"user_behavior"`
	// 系统通知配置（示例）
	Notification KafkaConfig `yaml:"notification"`
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		utils.Fatal("Failed to read config file", zap.Error(err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		utils.Fatal("Failed to parse config", zap.Error(err))
	}

	return &cfg
}

// local使用本地config.yaml文件
// 其他为从配置中心获取配置
func GetConfigLocation() string {
	return os.Getenv("CONFIG_LOCATION")
}

// ParseConfig 解析配置内容（支持json/yaml）
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}
	return nil, fmt.Errorf("failed to parse config as json or yaml")
}
