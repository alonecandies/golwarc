package cache

import "time"

// CacheClient defines the interface for cache operations
// This enables mocking in tests and provides a consistent API across different cache implementations
type CacheClient interface {
	// Get retrieves a value from the cache
	Get(key string) (string, error)

	// Set stores a value in the cache with a TTL
	Set(key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from the cache
	Delete(key string) error

	// Exists checks if a key exists in the cache
	Exists(key string) (bool, error)

	// Close closes the cache connection
	Close() error

	// Ping checks if the cache connection is alive
	Ping() error
}

// JSONCacheClient extends CacheClient with JSON serialization support
type JSONCacheClient interface {
	CacheClient

	// GetJSON retrieves a JSON value and unmarshals it into dest
	GetJSON(key string, dest interface{}) error

	// SetJSON stores a value as JSON in the cache
	SetJSON(key string, value interface{}, ttl time.Duration) error
}

// Ensure RedisClient implements the JSONCacheClient interface
var _ JSONCacheClient = (*RedisClient)(nil)
