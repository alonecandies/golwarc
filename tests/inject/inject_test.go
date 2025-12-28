package inject_test

import (
	"os"
	"testing"

	"github.com/alonecandies/golwarc/inject"
)

// TestNewContainer tests DI container creation with valid config
func TestNewContainer(t *testing.T) {
	// Create a minimal test config file
	configContent := `
app:
  name: test-golwarc
  environment: test
  port: 8080

logger:
  level: info
  development: true
  output_paths:
    - stdout

cache:
  lru:
    size: 100
  redis:
    addr: ""

database:
  mysql:
    host: ""
  postgresql:
    host: ""
  clickhouse:
    host: ""

message_queue:
  kafka:
    brokers: []
  rabbitmq:
    url: ""
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, writeErr := tmpFile.WriteString(configContent); writeErr != nil {
		t.Fatalf("Failed to write config: %v", writeErr)
	}
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container == nil {
		t.Fatal("Container should not be nil")
	}

	if container.Logger == nil {
		t.Error("Logger should be initialized")
	}

	if container.Config == nil {
		t.Error("Config should be initialized")
	}
}

// TestNewContainerWithInvalidConfig tests handling of invalid config
func TestNewContainerWithInvalidConfig(t *testing.T) {
	container, err := inject.NewContainer("nonexistent-file.yaml")

	// Should still create container with defaults
	if err != nil {
		t.Fatalf("Container creation should not fail: %v", err)
	}
	defer container.Close()

	if container.Logger == nil {
		t.Error("Logger should still be initialized with defaults")
	}
}

// TestContainerLRUCacheInitialization tests LRU cache conditional init
func TestContainerLRUCacheInitialization(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 1000
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container.LRUCache == nil {
		t.Error("LRU cache should be initialized when size > 0")
	}
}

// TestContainerLRUCacheNotInitialized tests that LRU cache is NOT initialized when size is 0
func TestContainerLRUCacheNotInitialized(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 0
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container.LRUCache != nil {
		t.Error("LRU cache should NOT be initialized when size is 0")
	}
}

// TestContainerClose tests cleanup functionality
func TestContainerClose(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 100
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Close should not error
	if err := container.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// TestContainerCloseMultipleTimes tests that Close is idempotent
func TestContainerCloseMultipleTimes(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 100
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// First close
	if err := container.Close(); err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Second close should also not panic
	if err := container.Close(); err != nil {
		// May error, but shouldn't panic
		t.Logf("Second Close() error (acceptable): %v", err)
	}
}

// TestContainerHealth tests health check functionality
func TestContainerHealth(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 100
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	health := container.Health()

	if health == nil {
		t.Fatal("Health should return a map")
	}

	// Logger should always be healthy
	if loggerHealth, ok := health["logger"]; !ok || !loggerHealth {
		t.Error("Logger should be healthy")
	}

	// Config should always be healthy
	if configHealth, ok := health["config"]; !ok || !configHealth {
		t.Error("Config should be healthy")
	}

	// LRU cache should be healthy when initialized
	if lruHealth, ok := health["lru_cache"]; !ok || !lruHealth {
		t.Error("LRU cache should be healthy when initialized")
	}
}

// TestContainerHealthWithoutServices tests health when services not configured
func TestContainerHealthWithoutServices(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 0
  redis:
    addr: ""
database:
  mysql:
    host: ""
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	health := container.Health()

	// Services not configured should be marked as unhealthy
	if health["redis"] {
		t.Error("Redis should not be healthy when not configured")
	}
	if health["mysql"] {
		t.Error("MySQL should not be healthy when not configured")
	}
	if health["postgresql"] {
		t.Error("PostgreSQL should not be healthy when not configured")
	}
	if health["clickhouse"] {
		t.Error("ClickHouse should not be healthy when not configured")
	}

	// LRU cache should not be healthy when size is 0
	if health["lru_cache"] {
		t.Error("LRU cache should not be healthy when size is 0")
	}
}

// TestContainerConfigValues tests that config values are properly set
func TestContainerConfigValues(t *testing.T) {
	configContent := `
app:
  name: test-app
  environment: test
  port: 9000
logger:
  level: debug
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container.Config.App.Name != "test-app" {
		t.Errorf("App name = %v, want test-app", container.Config.App.Name)
	}
	if container.Config.App.Port != 9000 {
		t.Errorf("App port = %v, want 9000", container.Config.App.Port)
	}
	if container.Config.Logger.Level != "debug" {
		t.Errorf("Logger level = %v, want debug", container.Config.Logger.Level)
	}
}

// TestContainerKafkaDefaultTopic tests Kafka initialization with default topic
func TestContainerKafkaDefaultTopic(t *testing.T) {
	configContent := `
logger:
  level: info
message_queue:
  kafka:
    brokers:
      - "localhost:9092"
    topic: ""
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	// Kafka should be initialized even though connection may fail
	if container.KafkaClient == nil {
		t.Error("Kafka client should be initialized")
	}

	// Verify it's in health status
	health := container.Health()
	if !health["kafka"] {
		t.Error("Kafka should be marked as initialized in health")
	}
}

// TestContainerKafkaCustomTopic tests Kafka initialization with custom topic
func TestContainerKafkaCustomTopic(t *testing.T) {
	configContent := `
logger:
  level: info
message_queue:
  kafka:
    brokers:
      - "localhost:9092"
    topic: "custom-topic"
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container.KafkaClient == nil {
		t.Error("Kafka client should be initialized")
	}
}

// TestContainerNoKafka tests that Kafka is not initialized without brokers
func TestContainerNoKafka(t *testing.T) {
	configContent := `
logger:
  level: info
message_queue:
  kafka:
    brokers: []
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	if container.KafkaClient != nil {
		t.Error("Kafka client should NOT be initialized without brokers")
	}

	health := container.Health()
	if health["kafka"] {
		t.Error("Kafka should not be healthy without brokers")
	}
}

// TestContainerPartialInitialization tests graceful degradation
func TestContainerPartialInitialization(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 100
  redis:
    addr: "invalid-redis-host:6379"
database:
  mysql:
    host: "invalid-mysql-host"
    port: 3306
    user: "test"
    password: "test"
    database: "test"
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Container creation should succeed even if some services fail: %v", err)
	}
	defer container.Close()

	// LRU should work
	if container.LRUCache == nil {
		t.Error("LRU cache should be initialized")
	}

	// Redis and MySQL should fail gracefully (nil)
	if container.RedisClient != nil {
		t.Log("Redis client unexpectedly initialized (connection may exist)")
	}
	if container.MySQLClient != nil {
		t.Log("MySQL client unexpectedly initialized (connection may exist)")
	}
}

// TestContainerAllServicesDisabled tests container with minimal config
func TestContainerAllServicesDisabled(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 0
  redis:
    addr: ""
database:
  mysql:
    host: ""
  postgresql:
    host: ""
  clickhouse:
    host: ""
message_queue:
  kafka:
    brokers: []
  rabbitmq:
    url: ""
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	// Logger and Config should always be present
	if container.Logger == nil {
		t.Error("Logger should always be initialized")
	}
	if container.Config == nil {
		t.Error("Config should always be initialized")
	}

	// All optional services should be nil
	if container.LRUCache != nil {
		t.Error("LRU cache should be nil")
	}
	if container.RedisClient != nil {
		t.Error("Redis client should be nil")
	}
	if container.MySQLClient != nil {
		t.Error("MySQL client should be nil")
	}
	if container.PGClient != nil {
		t.Error("PostgreSQL client should be nil")
	}
	if container.CHClient != nil {
		t.Error("ClickHouse client should be nil")
	}
	if container.KafkaClient != nil {
		t.Error("Kafka client should be nil")
	}
	if container.RabbitClient != nil {
		t.Error("RabbitMQ client should be nil")
	}
}

// TestContainerHealthAllServices tests health with all service types
func TestContainerHealthAllServices(t *testing.T) {
	configContent := `
logger:
  level: info
cache:
  lru:
    size: 100
`

	tmpFile, err := os.CreateTemp("", "inject-config-*.yaml")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	container, err := inject.NewContainer(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	health := container.Health()

	// Check that all expected keys are present
	expectedKeys := []string{
		"logger", "config", "lru_cache",
		"redis", "mysql", "postgresql", "clickhouse",
		"kafka", "rabbitmq",
	}

	for _, key := range expectedKeys {
		if _, ok := health[key]; !ok {
			t.Errorf("Health map missing key: %s", key)
		}
	}
}
