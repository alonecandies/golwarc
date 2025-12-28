package messagequeue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient wraps RabbitMQ operations
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	URL string
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(config RabbitMQConfig) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close() // Best effort cleanup
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		url:     config.URL,
	}, nil
}

// DeclareQueue declares a queue
func (r *RabbitMQClient) DeclareQueue(name string, durable bool) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
}

// DeclareQueueWithArgs declares a queue with custom arguments
func (r *RabbitMQClient) DeclareQueueWithArgs(name string, durable bool, args amqp.Table) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments
	)
}

// DeclareExchange declares an exchange
func (r *RabbitMQClient) DeclareExchange(name, kind string) error {
	return r.channel.ExchangeDeclare(
		name,  // name
		kind,  // type (direct, fanout, topic, headers)
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// BindQueue binds a queue to an exchange with a routing key
func (r *RabbitMQClient) BindQueue(queueName, exchangeName, routingKey string) error {
	return r.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
}

// Publish publishes a message to a queue
func (r *RabbitMQClient) Publish(ctx context.Context, queue string, message []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
			Timestamp:   time.Now(),
		},
	)
}

// PublishToExchange publishes a message to an exchange with a routing key
func (r *RabbitMQClient) PublishToExchange(ctx context.Context, exchange, routingKey string, message []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
			Timestamp:   time.Now(),
		},
	)
}

// PublishWithHeaders publishes a message with custom headers
func (r *RabbitMQClient) PublishWithHeaders(ctx context.Context, queue string, message []byte, headers amqp.Table) error {
	return r.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
			Headers:     headers,
			Timestamp:   time.Now(),
		},
	)
}

// Consume consumes messages from a queue
func (r *RabbitMQClient) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			if err := handler(msg.Body); err != nil {
				// Negative acknowledgment - requeue the message
				_ = msg.Nack(false, true) // Error intentionally ignored
				return fmt.Errorf("handler error: %w", err)
			}

			// Acknowledge the message
			_ = msg.Ack(false) // Error intentionally ignored
		}
	}
}

// ConsumeWithManualAck consumes messages with manual acknowledgment control
func (r *RabbitMQClient) ConsumeWithManualAck(queue string) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

// SetQoS sets the Quality of Service for the channel
func (r *RabbitMQClient) SetQoS(prefetchCount, prefetchSize int, global bool) error {
	return r.channel.Qos(prefetchCount, prefetchSize, global)
}

// PurgeQueue removes all messages from a queue
func (r *RabbitMQClient) PurgeQueue(queue string) (int, error) {
	return r.channel.QueuePurge(queue, false)
}

// DeleteQueue deletes a queue
func (r *RabbitMQClient) DeleteQueue(queue string) error {
	_, err := r.channel.QueueDelete(queue, false, false, false)
	return err
}

// GetChannel returns the underlying AMQP channel for advanced operations
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQClient) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}

// IsClosed checks if the connection is closed
func (r *RabbitMQClient) IsClosed() bool {
	return r.conn.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (r *RabbitMQClient) Reconnect() error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close() // Best effort cleanup
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Close old connection if still open
	if !r.conn.IsClosed() {
		_ = r.channel.Close() // Best effort cleanup
		_ = r.conn.Close()    // Best effort cleanup
	}

	r.conn = conn
	r.channel = channel

	return nil
}
