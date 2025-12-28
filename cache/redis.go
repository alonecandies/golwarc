package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alonecandies/golwarc/libs"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with common operations.
// It provides a simplified interface for Redis operations with JSON support,
// automatic error handling, and TLS configuration support.
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig holds Redis connection configuration.
// Addr: Redis server address (host:port)
// Password: Authentication password (empty for no auth)
// DB: Database number (0-15)
// TLS: Optional TLS configuration for secure connections
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TLS      *libs.TLSConfig
}

// NewRedisClient creates a new Redis client with the provided configuration.
// It establishes a connection to the Redis server and tests connectivity with a ping.
// Returns an error if the connection fails or TLS configuration is invalid.
//
// Example:
//
//	client, err := NewRedisClient(RedisConfig{
//	    Addr: "localhost:6379",
//	    Password: "",
//	    DB: 0,
//	})
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	options := &redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	}

	// Configure TLS if enabled
	if config.TLS != nil && config.TLS.Enabled {
		tlsConfig, err := libs.CreateTLSConfig(config.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		options.TLSConfig = tlsConfig
	}

	client := redis.NewClient(options)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("key does not exist")
	}
	return val, err
}

// GetJSON retrieves a JSON value and unmarshals it
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	val, err := r.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in Redis with optional TTL
func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

// SetJSON stores a JSON value in Redis
func (r *RedisClient) SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(key, data, ttl)
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// DeleteMany removes multiple keys from Redis
func (r *RedisClient) DeleteMany(keys ...string) error {
	return r.client.Del(r.ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	return result > 0, err
}

// Expire sets an expiration on a key
func (r *RedisClient) Expire(key string, ttl time.Duration) error {
	return r.client.Expire(r.ctx, key, ttl).Err()
}

// TTL gets the remaining time to live for a key
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

// GetWithTTL retrieves a value along with its remaining TTL
func (r *RedisClient) GetWithTTL(key string) (string, time.Duration, error) {
	val, err := r.Get(key)
	if err != nil {
		return "", 0, err
	}

	ttl, err := r.TTL(key)
	if err != nil {
		return val, 0, err
	}

	return val, ttl, nil
}

// Increment increments a key's value by 1
func (r *RedisClient) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// IncrementBy increments a key's value by the specified amount
func (r *RedisClient) IncrementBy(key string, value int64) (int64, error) {
	return r.client.IncrBy(r.ctx, key, value).Result()
}

// Decrement decrements a key's value by 1
func (r *RedisClient) Decrement(key string) (int64, error) {
	return r.client.Decr(r.ctx, key).Result()
}

// SetNX sets a key only if it doesn't exist (atomic)
func (r *RedisClient) SetNX(key string, value interface{}, ttl time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, ttl).Result()
}

// MGet retrieves multiple keys at once
func (r *RedisClient) MGet(keys ...string) ([]interface{}, error) {
	return r.client.MGet(r.ctx, keys...).Result()
}

// MSet sets multiple keys at once
func (r *RedisClient) MSet(pairs map[string]interface{}) error {
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
	}
	return r.client.MSet(r.ctx, args...).Err()
}

// FlushDB clears all keys in the current database
func (r *RedisClient) FlushDB() error {
	return r.client.FlushDB(r.ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping checks if the Redis connection is alive
func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
