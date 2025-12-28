package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Logger       LoggerConfig       `mapstructure:"logger"`
	Cache        CacheConfig        `mapstructure:"cache"`
	Database     DatabaseConfig     `mapstructure:"database"`
	MessageQueue MessageQueueConfig `mapstructure:"message_queue"`
	Temporal     TemporalConfig     `mapstructure:"temporal"`
	Crawler      CrawlerConfig      `mapstructure:"crawler"`
}

// AppConfig holds general application settings
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level       string   `mapstructure:"level"`
	Development bool     `mapstructure:"development"`
	OutputPaths []string `mapstructure:"output_paths"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	LRU   LRUConfig   `mapstructure:"lru"`
	Redis RedisConfig `mapstructure:"redis"`
}

// LRUConfig holds LRU cache settings
type LRUConfig struct {
	Size int `mapstructure:"size"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string    `mapstructure:"addr"`
	Password string    `mapstructure:"password"`
	DB       int       `mapstructure:"db"`
	TLS      TLSConfig `mapstructure:"tls"`
}

// DatabaseConfig holds database configurations
type DatabaseConfig struct {
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	PostgreSQL PostgreSQLConfig `mapstructure:"postgresql"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	BigTable   BigTableConfig   `mapstructure:"bigtable"`
}

// MySQLConfig holds MySQL connection settings
type MySQLConfig struct {
	Host     string    `mapstructure:"host"`
	Port     int       `mapstructure:"port"`
	User     string    `mapstructure:"user"`
	Password string    `mapstructure:"password"`
	Database string    `mapstructure:"database"`
	Charset  string    `mapstructure:"charset"`
	TLS      TLSConfig `mapstructure:"tls"`
}

// PostgreSQLConfig holds PostgreSQL connection settings
type PostgreSQLConfig struct {
	Host     string    `mapstructure:"host"`
	Port     int       `mapstructure:"port"`
	User     string    `mapstructure:"user"`
	Password string    `mapstructure:"password"`
	Database string    `mapstructure:"database"`
	SSLMode  string    `mapstructure:"sslmode"`
	TimeZone string    `mapstructure:"timezone"`
	TLS      TLSConfig `mapstructure:"tls"`
}

// ClickHouseConfig holds ClickHouse connection settings
type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// BigTableConfig holds BigTable connection settings
type BigTableConfig struct {
	ProjectID  string `mapstructure:"project_id"`
	InstanceID string `mapstructure:"instance_id"`
}

// MessageQueueConfig holds message queue configurations
type MessageQueueConfig struct {
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// KafkaConfig holds Kafka connection settings
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
	Topic   string   `mapstructure:"topic"`
}

// RabbitMQConfig holds RabbitMQ connection settings
type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

// TemporalConfig holds Temporal connection settings
type TemporalConfig struct {
	HostPort  string `mapstructure:"host_port"`
	Namespace string `mapstructure:"namespace"`
}

// CrawlerConfig holds crawler settings
type CrawlerConfig struct {
	UserAgent         string          `mapstructure:"user_agent"`
	MaxDepth          int             `mapstructure:"max_depth"`
	Concurrency       int             `mapstructure:"concurrency"`
	RequestTimeout    int             `mapstructure:"request_timeout"`
	RateLimitDelay    int             `mapstructure:"rate_limit_delay"`
	SeleniumURL       string          `mapstructure:"selenium_url"`
	PlaywrightBrowser string          `mapstructure:"playwright_browser"`
	RateLimit         RateLimitConfig `mapstructure:"rate_limit"`
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigFile(path)

	// Enable environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// TLSConfig holds TLS/SSL configuration
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	CACert             string `mapstructure:"ca_cert"`
	ClientCert         string `mapstructure:"client_cert"`
	ClientKey          string `mapstructure:"client_key"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	Delay          int  `mapstructure:"delay"`            // milliseconds
	RandomDelay    int  `mapstructure:"random_delay"`     // milliseconds
	MaxConcurrent  int  `mapstructure:"max_concurrent"`   // max concurrent requests
	RequestsPerSec int  `mapstructure:"requests_per_sec"` // max requests per second
}

// LoadConfigOrDefault loads config from file or returns default config
func LoadConfigOrDefault(path string) *Config {
	config, err := LoadConfig(path)
	if err != nil {
		return GetDefaultConfig()
	}
	return config
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "golwarc",
			Environment: "development",
			Port:        8080,
		},
		Logger: LoggerConfig{
			Level:       "info",
			Development: true,
			OutputPaths: []string{"stdout"},
		},
		Cache: CacheConfig{
			LRU: LRUConfig{
				Size: 1000,
			},
			Redis: RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
		},
		Crawler: CrawlerConfig{
			UserAgent:         "Mozilla/5.0 (compatible; GolwarcBot/1.0)",
			MaxDepth:          3,
			Concurrency:       5,
			RequestTimeout:    30,
			RateLimitDelay:    1000,
			SeleniumURL:       "http://localhost:4444/wd/hub",
			PlaywrightBrowser: "chromium",
		},
	}
}
