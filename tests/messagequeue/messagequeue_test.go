package messagequeue_test

import (
	"context"
	"testing"
	"time"

	messagequeue "github.com/alonecandies/golwarc/message-queue"
	"github.com/segmentio/kafka-go"
)

// =============================================================================
// Kafka Producer Configuration Tests
// =============================================================================

func TestKafkaProducerConfig(t *testing.T) {
	tests := []struct {
		name   string
		config messagequeue.KafkaProducerConfig
	}{
		{
			name: "single broker",
			config: messagequeue.KafkaProducerConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			},
		},
		{
			name: "multiple brokers",
			config: messagequeue.KafkaProducerConfig{
				Brokers: []string{"broker1:9092", "broker2:9092", "broker3:9092"},
				Topic:   "multi-topic",
			},
		},
		{
			name: "empty brokers",
			config: messagequeue.KafkaProducerConfig{
				Brokers: []string{},
				Topic:   "test-topic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that config can be created
			if tt.config.Topic == "" && tt.name != "empty brokers" {
				t.Error("Topic should not be empty")
			}
		})
	}
}

func TestNewKafkaProducer(t *testing.T) {
	tests := []struct {
		name   string
		config messagequeue.KafkaProducerConfig
	}{
		{
			name: "valid single broker",
			config: messagequeue.KafkaProducerConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			},
		},
		{
			name: "valid multiple brokers",
			config: messagequeue.KafkaProducerConfig{
				Brokers: []string{"broker1:9092", "broker2:9092"},
				Topic:   "multi-topic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := messagequeue.NewKafkaProducer(tt.config)
			if producer == nil {
				t.Fatal("Producer should not be nil")
			}
			defer producer.Close()
		})
	}
}

func TestKafkaProducer_Close(t *testing.T) {
	producer := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	})

	err := producer.Close()
	// Close should not panic even if connection was never established
	if err != nil {
		t.Logf("Close() error (acceptable): %v", err)
	}

	// Multiple closes should not panic
	err = producer.Close()
	if err != nil {
		t.Logf("Second Close() error (acceptable): %v", err)
	}
}

// =============================================================================
// Kafka Consumer Configuration Tests
// =============================================================================

func TestKafkaConsumerConfig(t *testing.T) {
	tests := []struct {
		name   string
		config messagequeue.KafkaConsumerConfig
	}{
		{
			name: "valid config",
			config: messagequeue.KafkaConsumerConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "test-group",
			},
		},
		{
			name: "multiple brokers",
			config: messagequeue.KafkaConsumerConfig{
				Brokers: []string{"broker1:9092", "broker2:9092"},
				Topic:   "multi-topic",
				GroupID: "consumer-group",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.config.Brokers) == 0 {
				t.Error("Brokers should not be empty")
			}
			if tt.config.Topic == "" {
				t.Error("Topic should not be empty")
			}
			if tt.config.GroupID == "" {
				t.Error("GroupID should not be empty")
			}
		})
	}
}

func TestNewKafkaConsumer(t *testing.T) {
	tests := []struct {
		name   string
		config messagequeue.KafkaConsumerConfig
	}{
		{
			name: "valid config",
			config: messagequeue.KafkaConsumerConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "test-group",
			},
		},
		{
			name: "empty group ID",
			config: messagequeue.KafkaConsumerConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
				GroupID: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := messagequeue.NewKafkaConsumer(tt.config)
			if consumer == nil {
				t.Fatal("Consumer should not be nil")
			}
			defer consumer.Close()
		})
	}
}

func TestKafkaConsumer_Close(t *testing.T) {
	consumer := messagequeue.NewKafkaConsumer(messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
	})

	err := consumer.Close()
	// Close should not panic even if connection was never established
	if err != nil {
		t.Logf("Close() error (acceptable): %v", err)
	}
}

func TestKafkaConsumer_Stats(t *testing.T) {
	consumer := messagequeue.NewKafkaConsumer(messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
	})
	defer consumer.Close()

	// Stats should be callable even without connection
	stats := consumer.Stats()
	_ = stats // Just ensure it doesn't panic
}

// =============================================================================
// Kafka Producer Integration Tests (Skipped without Kafka)
// =============================================================================

func TestKafkaProducer_Produce_Integration(t *testing.T) {
	producer := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	})
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := producer.Produce(ctx, []byte("key"), []byte("test message"))
	if err != nil {
		t.Skip("Kafka not available:", err)
	}
}

func TestKafkaProducer_ProduceWithHeaders_Integration(t *testing.T) {
	producer := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	})
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Request-ID": "12345",
	}

	err := producer.ProduceWithHeaders(ctx, []byte("key"), []byte("test message"), headers)
	if err != nil {
		t.Skip("Kafka not available:", err)
	}
}

func TestKafkaProducer_ProduceBatch_Integration(t *testing.T) {
	producer := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	})
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := []kafka.Message{
		{Key: []byte("key1"), Value: []byte("message1")},
		{Key: []byte("key2"), Value: []byte("message2")},
		{Key: []byte("key3"), Value: []byte("message3")},
	}

	err := producer.ProduceBatch(ctx, messages)
	if err != nil {
		t.Skip("Kafka not available:", err)
	}
}

func TestKafkaConsumer_ReadMessage_Integration(t *testing.T) {
	consumer := messagequeue.NewKafkaConsumer(messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
	})
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := consumer.ReadMessage(ctx)
	if err != nil {
		t.Skip("Kafka not available or no messages:", err)
	}
}

func TestKafkaConsumer_FetchMessage_Integration(t *testing.T) {
	consumer := messagequeue.NewKafkaConsumer(messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
	})
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := consumer.FetchMessage(ctx)
	if err != nil {
		t.Skip("Kafka not available or no messages:", err)
	}
}

func TestKafkaConsumer_SetOffset_Integration(t *testing.T) {
	consumer := messagequeue.NewKafkaConsumer(messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
	})
	defer consumer.Close()

	err := consumer.SetOffset(0)
	if err != nil {
		t.Skip("Kafka not available:", err)
	}
}

// =============================================================================
// RabbitMQ Configuration Tests
// =============================================================================

func TestRabbitMQConfig(t *testing.T) {
	tests := []struct {
		name   string
		config messagequeue.RabbitMQConfig
	}{
		{
			name: "valid local config",
			config: messagequeue.RabbitMQConfig{
				URL: "amqp://guest:guest@localhost:5672/",
			},
		},
		{
			name: "remote config",
			config: messagequeue.RabbitMQConfig{
				URL: "amqp://user:pass@rabbitmq.example.com:5672/vhost",
			},
		},
		{
			name: "empty URL",
			config: messagequeue.RabbitMQConfig{
				URL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Config can be created with any value
			_ = tt.config
		})
	}
}

func TestNewRabbitMQClient_InvalidURL(t *testing.T) {
	_, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "invalid://url",
	})
	if err == nil {
		t.Error("Should return error for invalid URL")
	}
}

func TestNewRabbitMQClient_EmptyURL(t *testing.T) {
	_, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "",
	})
	if err == nil {
		t.Error("Should return error for empty URL")
	}
}

// =============================================================================
// RabbitMQ Integration Tests (Skipped without RabbitMQ)
// =============================================================================

func TestNewRabbitMQClient_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestRabbitMQClient_DeclareQueue_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer client.Close()

	_, err = client.DeclareQueue("test-queue-unit", true)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	// Cleanup
	client.DeleteQueue("test-queue-unit")
}

func TestRabbitMQClient_Publish_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer client.Close()

	client.DeclareQueue("test-publish-queue", false)
	defer client.DeleteQueue("test-publish-queue")

	ctx := context.Background()
	err = client.Publish(ctx, "test-publish-queue", []byte("test message"))
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}
}

func TestRabbitMQClient_DeclareExchange_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer client.Close()

	err = client.DeclareExchange("test-exchange-unit", "direct")
	if err != nil {
		t.Fatalf("Failed to declare exchange: %v", err)
	}
}

func TestRabbitMQClient_BindQueue_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer client.Close()

	client.DeclareExchange("test-bind-exchange", "direct")
	client.DeclareQueue("test-bind-queue", false)
	defer client.DeleteQueue("test-bind-queue")

	err = client.BindQueue("test-bind-queue", "test-bind-exchange", "test-key")
	if err != nil {
		t.Fatalf("Failed to bind queue: %v", err)
	}
}

func TestRabbitMQClient_IsClosed_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}

	if client.IsClosed() {
		t.Error("Connection should not be closed initially")
	}

	client.Close()

	if !client.IsClosed() {
		t.Error("Connection should be closed after Close()")
	}
}

func TestRabbitMQClient_Close_Multiple_Integration(t *testing.T) {
	client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}

	// First close
	err = client.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Second close should not panic
	err = client.Close()
	// May return error, but shouldn't panic
	_ = err
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewKafkaProducer(b *testing.B) {
	config := messagequeue.KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "benchmark-topic",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer := messagequeue.NewKafkaProducer(config)
		producer.Close()
	}
}

func BenchmarkNewKafkaConsumer(b *testing.B) {
	config := messagequeue.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "benchmark-topic",
		GroupID: "benchmark-group",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		consumer := messagequeue.NewKafkaConsumer(config)
		consumer.Close()
	}
}
