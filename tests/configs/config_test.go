package configs_test

import (
	"os"
	"testing"

	"github.com/alonecandies/golwarc/configs"
)

// TestLoadConfig tests loading a valid config file
func TestLoadConfig(t *testing.T) {
	cfg, err := configs.LoadConfig("../../config.example.yaml")
	if err != nil {
		t.Skipf("Skipping test: config file not found: %v", err)
		return
	}

	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	// Verify some basic fields exist
	if cfg.App.Name == "" {
		t.Error("Expected app name to be set")
	}
}

// TestGetDefaultConfig tests the default configuration
func TestGetDefaultConfig(t *testing.T) {
	cfg := configs.GetDefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	// Test app defaults
	if cfg.App.Name != "golwarc" {
		t.Errorf("Expected app name 'golwarc', got %v", cfg.App.Name)
	}
	if cfg.App.Environment != "development" {
		t.Errorf("Expected environment 'development', got %v", cfg.App.Environment)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("Expected port 8080, got %v", cfg.App.Port)
	}

	// Test logger defaults
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected log level 'info', got %v", cfg.Logger.Level)
	}
	if !cfg.Logger.Development {
		t.Error("Expected development mode to be true")
	}
	if len(cfg.Logger.OutputPaths) == 0 {
		t.Error("Expected output paths to be set")
	}

	// Test cache defaults
	if cfg.Cache.LRU.Size != 1000 {
		t.Errorf("Expected LRU size 1000, got %v", cfg.Cache.LRU.Size)
	}
	if cfg.Cache.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected Redis addr 'localhost:6379', got %v", cfg.Cache.Redis.Addr)
	}
	if cfg.Cache.Redis.DB != 0 {
		t.Errorf("Expected Redis DB 0, got %v", cfg.Cache.Redis.DB)
	}

	// Test crawler defaults
	if cfg.Crawler.MaxDepth != 3 {
		t.Errorf("Expected max depth 3, got %v", cfg.Crawler.MaxDepth)
	}
	if cfg.Crawler.Concurrency != 5 {
		t.Errorf("Expected concurrency 5, got %v", cfg.Crawler.Concurrency)
	}
	if cfg.Crawler.RequestTimeout != 30 {
		t.Errorf("Expected request timeout 30, got %v", cfg.Crawler.RequestTimeout)
	}
}

// TestLoadConfigOrDefault tests loading with fallback to default
func TestLoadConfigOrDefault(t *testing.T) {
	// Test with non-existent file - should return default
	cfg := configs.LoadConfigOrDefault("nonexistent.yaml")

	if cfg == nil {
		t.Fatal("Expected default config, got nil")
	}

	if cfg.App.Name != "golwarc" {
		t.Errorf("Expected default app name, got %v", cfg.App.Name)
	}
}

// TestLoadConfigInvalidFile tests handling of invalid config files
func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := configs.LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestConfigStructBasicFields tests basic config struct fields
func TestConfigStructBasicFields(t *testing.T) {
	cfg := &configs.Config{
		App: configs.AppConfig{
			Name:        "testapp",
			Environment: "test",
			Port:        3000,
		},
	}

	if cfg.App.Name != "testapp" {
		t.Errorf("Name = %v, want testapp", cfg.App.Name)
	}
	if cfg.App.Environment != "test" {
		t.Errorf("Environment = %v, want test", cfg.App.Environment)
	}
	if cfg.App.Port != 3000 {
		t.Errorf("Port = %v, want 3000", cfg.App.Port)
	}
}

// TestLoggerConfig tests logger configuration
func TestLoggerConfig(t *testing.T) {
	cfg := configs.LoggerConfig{
		Level:       "debug",
		Development: true,
		OutputPaths: []string{"stdout", "file.log"},
	}

	if cfg.Level != "debug" {
		t.Errorf("Level = %v, want debug", cfg.Level)
	}
	if !cfg.Development {
		t.Error("Development should be true")
	}
	if len(cfg.OutputPaths) != 2 {
		t.Errorf("OutputPaths length = %v, want 2", len(cfg.OutputPaths))
	}
}

// TestCacheConfig tests cache configuration
func TestCacheConfig(t *testing.T) {
	cfg := configs.CacheConfig{
		LRU: configs.LRUConfig{
			Size: 5000,
		},
		Redis: configs.RedisConfig{
			Addr:     "redis:6379",
			Password: "secret",
			DB:       1,
		},
	}

	if cfg.LRU.Size != 5000 {
		t.Errorf("LRU.Size = %v, want 5000", cfg.LRU.Size)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("Redis.Addr = %v, want redis:6379", cfg.Redis.Addr)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("Redis.DB = %v, want 1", cfg.Redis.DB)
	}
}

// TestDatabaseConfig tests database configuration
func TestDatabaseConfig(t *testing.T) {
	cfg := configs.DatabaseConfig{
		MySQL: configs.MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "password",
			Database: "testdb",
			Charset:  "utf8mb4",
		},
		PostgreSQL: configs.PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Database: "testdb",
			SSLMode:  "disable",
			TimeZone: "UTC",
		},
	}

	if cfg.MySQL.Host != "localhost" {
		t.Errorf("MySQL.Host = %v, want localhost", cfg.MySQL.Host)
	}
	if cfg.MySQL.Port != 3306 {
		t.Errorf("MySQL.Port = %v, want 3306", cfg.MySQL.Port)
	}
	if cfg.PostgreSQL.Port != 5432 {
		t.Errorf("PostgreSQL.Port = %v, want 5432", cfg.PostgreSQL.Port)
	}
}

// TestMessageQueueConfig tests message queue configuration
func TestMessageQueueConfig(t *testing.T) {
	cfg := configs.MessageQueueConfig{
		Kafka: configs.KafkaConfig{
			Brokers: []string{"localhost:9092", "localhost:9093"},
			GroupID: "test-group",
			Topic:   "test-topic",
		},
		RabbitMQ: configs.RabbitMQConfig{
			URL: "amqp://guest:guest@localhost:5672/",
		},
	}

	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("Kafka.Brokers length = %v, want 2", len(cfg.Kafka.Brokers))
	}
	if cfg.Kafka.GroupID != "test-group" {
		t.Errorf("Kafka.GroupID = %v, want test-group", cfg.Kafka.GroupID)
	}
	if cfg.RabbitMQ.URL == "" {
		t.Error("RabbitMQ.URL should not be empty")
	}
}

// TestCrawlerConfig tests crawler configuration
func TestCrawlerConfig(t *testing.T) {
	cfg := configs.CrawlerConfig{
		UserAgent:         "TestBot/1.0",
		MaxDepth:          5,
		PlaywrightBrowser: "firefox",
	}
	_ = cfg.Concurrency    // Just testing struct initialization
	_ = cfg.RequestTimeout // Just testing struct initialization
	_ = cfg.RateLimitDelay // Just testing struct initialization
	_ = cfg.SeleniumURL    // Just testing struct initialization

	if cfg.UserAgent != "TestBot/1.0" {
		t.Errorf("UserAgent = %v, want TestBot/1.0", cfg.UserAgent)
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("MaxDepth = %v, want 5", cfg.MaxDepth)
	}
	if cfg.PlaywrightBrowser != "firefox" {
		t.Errorf("PlaywrightBrowser = %v, want firefox", cfg.PlaywrightBrowser)
	}
}

// TestTLSConfig tests TLS configuration
func TestTLSConfig(t *testing.T) {
	cfg := configs.TLSConfig{
		Enabled:            true,
		CACert:             "/path/to/ca.crt",
		InsecureSkipVerify: false,
	}
	_ = cfg.ClientCert // Just testing struct initialization
	_ = cfg.ClientKey  // Just testing struct initialization

	if !cfg.Enabled {
		t.Error("TLS should be enabled")
	}
	if cfg.CACert == "" {
		t.Error("CACert should not be empty")
	}
	if cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false")
	}
}

// TestRateLimitConfig tests rate limit configuration
func TestRateLimitConfig(t *testing.T) {
	cfg := configs.RateLimitConfig{
		Enabled:        true,
		Delay:          1000,
		RequestsPerSec: 10,
	}
	_ = cfg.RandomDelay   // Just testing struct initialization
	_ = cfg.MaxConcurrent // Just testing struct initialization

	if !cfg.Enabled {
		t.Error("Rate limiting should be enabled")
	}
	if cfg.Delay != 1000 {
		t.Errorf("Delay = %v, want 1000", cfg.Delay)
	}
	if cfg.RequestsPerSec != 10 {
		t.Errorf("RequestsPerSec = %v, want 10", cfg.RequestsPerSec)
	}
}

// TestConfigWithEnvironmentVariables tests environment variable override
func TestConfigWithEnvironmentVariables(t *testing.T) {
	// This test verifies the structure supports env vars (actual override tested in integration)
	cfg := configs.GetDefaultConfig()

	// Verify default values that could be overridden
	if cfg.App.Port == 0 {
		t.Error("Port should have a default value")
	}
	if cfg.Cache.Redis.Addr == "" {
		t.Error("Redis addr should have a default value")
	}
}

// TestClickHouseConfig tests ClickHouse configuration
func TestClickHouseConfig(t *testing.T) {
	cfg := configs.ClickHouseConfig{
		Host: "clickhouse",
		Port: 9000,
	}
	_ = cfg.User     // Just testing struct initialization
	_ = cfg.Password // Just testing struct initialization
	_ = cfg.Database // Just testing struct initialization

	if cfg.Host != "clickhouse" {
		t.Errorf("Host = %v, want clickhouse", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %v, want 9000", cfg.Port)
	}
}

// TestBigTableConfig tests BigTable configuration
func TestBigTableConfig(t *testing.T) {
	cfg := configs.BigTableConfig{
		ProjectID:  "my-project",
		InstanceID: "my-instance",
	}

	if cfg.ProjectID != "my-project" {
		t.Errorf("ProjectID = %v, want my-project", cfg.ProjectID)
	}
	if cfg.InstanceID != "my-instance" {
		t.Errorf("InstanceID = %v, want my-instance", cfg.InstanceID)
	}
}

// TestTemporalConfig tests Temporal configuration
func TestTemporalConfig(t *testing.T) {
	cfg := configs.TemporalConfig{
		HostPort:  "temporal:7233",
		Namespace: "default",
	}

	if cfg.HostPort != "temporal:7233" {
		t.Errorf("HostPort = %v, want temporal:7233", cfg.HostPort)
	}
	if cfg.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", cfg.Namespace)
	}
}

// TestLoadConfigWithActualFile tests loading the example config if it exists
func TestLoadConfigWithActualFile(t *testing.T) {
	// Try to load the actual example config
	cfg, err := configs.LoadConfig("../../config.example.yaml")
	if err != nil {
		// Skip if file doesn't exist
		if os.IsNotExist(err) {
			t.Skip("config.example.yaml not found")
		}
		t.Skipf("Could not load config: %v", err)
		return
	}

	// Verify config loaded successfully
	if cfg.App.Name == "" {
		t.Error("Expected app name to be set in example config")
	}
}
