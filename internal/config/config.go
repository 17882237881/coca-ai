package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	LLM    LLMConfig    `mapstructure:"llm"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Jaeger JaegerConfig `mapstructure:"jaeger"`
	Logger LoggerConfig `mapstructure:"logger"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider string `mapstructure:"provider"` // qwen, openai
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

// JaegerConfig Jaeger 配置
type JaegerConfig struct {
	AgentHost string `mapstructure:"agent_host"`
	AgentPort string `mapstructure:"agent_port"`
}

var globalConfig *Config

// InitConfig 初始化配置
func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/app/configs") // Docker 环境

	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}

	// 环境变量覆盖 (优先级: 环境变量 > 配置文件)
	if apiKey := os.Getenv("QWEN_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}
	if baseURL := os.Getenv("QWEN_BASE_URL"); baseURL != "" {
		cfg.LLM.BaseURL = baseURL
	}
	if model := os.Getenv("QWEN_MODEL"); model != "" {
		cfg.LLM.Model = model
	}
	// Kafka 环境变量覆盖
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = []string{brokers}
	}
	// Jaeger 环境变量覆盖
	if host := os.Getenv("JAEGER_AGENT_HOST"); host != "" {
		cfg.Jaeger.AgentHost = host
	}
	if port := os.Getenv("JAEGER_AGENT_PORT"); port != "" {
		cfg.Jaeger.AgentPort = port
	}

	globalConfig = &cfg
	return &cfg
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		return InitConfig()
	}
	return globalConfig
}
