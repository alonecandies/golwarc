package messagequeue

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps Kafka producer operations
type KafkaProducer struct {
	writer *kafka.Writer
}

// KafkaConsumer wraps Kafka consumer operations
type KafkaConsumer struct {
	reader *kafka.Reader
}

// KafkaProducerConfig holds Kafka producer configuration
type KafkaProducerConfig struct {
	Brokers []string
	Topic   string
}

// KafkaConsumerConfig holds Kafka consumer configuration
type KafkaConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(config KafkaProducerConfig) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(config KafkaConsumerConfig) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Brokers,
		Topic:          config.Topic,
		GroupID:        config.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})

	return &KafkaConsumer{
		reader: reader,
	}
}

// Produce sends a message to Kafka
func (p *KafkaProducer) Produce(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

// ProduceWithHeaders sends a message with headers to Kafka
func (p *KafkaProducer) ProduceWithHeaders(ctx context.Context, key, value []byte, headers map[string]string) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

// ProduceBatch sends multiple messages to Kafka
func (p *KafkaProducer) ProduceBatch(ctx context.Context, messages []kafka.Message) error {
	return p.writer.WriteMessages(ctx, messages...)
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// Consume reads messages from Kafka and processes them with the handler
func (c *KafkaConsumer) Consume(ctx context.Context, handler func(msg kafka.Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch message: %w", err)
			}

			if err := handler(msg); err != nil {
				return fmt.Errorf("handler error: %w", err)
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %w", err)
			}
		}
	}
}

// ReadMessage reads a single message from Kafka
func (c *KafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// FetchMessage fetches a message without committing
func (c *KafkaConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

// CommitMessages commits messages
func (c *KafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// SetOffset sets the consumer offset
func (c *KafkaConsumer) SetOffset(offset int64) error {
	return c.reader.SetOffset(offset)
}

// Stats returns consumer statistics
func (c *KafkaConsumer) Stats() kafka.ReaderStats {
	return c.reader.Stats()
}
