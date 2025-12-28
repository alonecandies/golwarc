package cache_test

import (
	"testing"
	"time"

	"github.com/alonecandies/golwarc/cache"
)

func TestLRUCache(t *testing.T) {
	lru, err := cache.NewLRUCache(2)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	// Test Set and Get
	lru.Set("key1", "value1")
	val, exists := lru.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test capacity
	lru.Set("key2", "value2")
	lru.Set("key3", "value3") // This should evict key1

	_, exists = lru.Get("key1")
	if exists {
		t.Error("Expected key1 to be evicted")
	}

	// Test Delete
	lru.Delete("key2")
	_, exists = lru.Get("key2")
	if exists {
		t.Error("Expected key2 to be deleted")
	}

	// Test Len
	if lru.Len() != 1 {
		t.Errorf("Expected length 1, got %d", lru.Len())
	}

	// Test Clear
	lru.Clear()
	if lru.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", lru.Len())
	}
}

func TestLRUCachePeek(t *testing.T) {
	lru, _ := cache.NewLRUCache(2)
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")

	// Peek should not update recency
	val, exists := lru.Peek("key1")
	if !exists || val != "value1" {
		t.Error("Peek failed")
	}

	// Add key3, which should evict key1 (not key2)
	lru.Set("key3", "value3")
	_, exists = lru.Get("key1")
	if exists {
		t.Error("Expected key1 to be evicted")
	}
}

func TestRedisClient(t *testing.T) {
	// Skip if Redis is not available
	client, err := cache.NewRedisClient(cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	// Test Set and Get
	key := "test:key"
	err = client.Set(key, "value", time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	val, err := client.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %s", val)
	}

	// Test Delete
	err = client.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	exists, err := client.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected key to be deleted")
	}
}

func TestRedisJSON(t *testing.T) {
	client, err := cache.NewRedisClient(cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	type TestStruct struct {
		Name  string
		Value int
	}

	data := TestStruct{Name: "test", Value: 42}
	key := "test:json"

	// Test SetJSON
	err = client.SetJSON(key, data, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set JSON: %v", err)
	}

	// Test GetJSON
	var result TestStruct
	err = client.GetJSON(key, &result)
	if err != nil {
		t.Fatalf("Failed to get JSON: %v", err)
	}

	if result.Name != data.Name || result.Value != data.Value {
		t.Errorf("Expected %+v, got %+v", data, result)
	}

	client.Delete(key)
}
