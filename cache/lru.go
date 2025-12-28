package cache

import (
	"errors"

	lru "github.com/hashicorp/golang-lru/v2"
)

// LRUCache wraps the hashicorp LRU cache with additional functionality
type LRUCache struct {
	cache *lru.Cache[string, interface{}]
}

// NewLRUCache creates a new LRU cache with the specified size
func NewLRUCache(size int) (*LRUCache, error) {
	if size <= 0 {
		return nil, errors.New("cache size must be positive")
	}

	cache, err := lru.New[string, interface{}](size)
	if err != nil {
		return nil, err
	}

	return &LRUCache{
		cache: cache,
	}, nil
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Set stores a value in the cache
func (c *LRUCache) Set(key string, value interface{}) bool {
	return c.cache.Add(key, value)
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) bool {
	return c.cache.Remove(key)
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.cache.Purge()
}

// Len returns the number of items in the cache
func (c *LRUCache) Len() int {
	return c.cache.Len()
}

// Contains checks if a key exists in the cache
func (c *LRUCache) Contains(key string) bool {
	return c.cache.Contains(key)
}

// Peek returns the value without updating the LRU
func (c *LRUCache) Peek(key string) (interface{}, bool) {
	return c.cache.Peek(key)
}

// Keys returns all keys in the cache (newest to oldest)
func (c *LRUCache) Keys() []string {
	return c.cache.Keys()
}

// Resize changes the cache size (evicts if necessary)
func (c *LRUCache) Resize(size int) (int, error) {
	if size <= 0 {
		return 0, errors.New("cache size must be positive")
	}
	return c.cache.Resize(size), nil
}
