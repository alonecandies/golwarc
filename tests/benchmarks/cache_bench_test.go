package benchmarks

import (
	"testing"
	"time"

	"github.com/alonecandies/golwarc/cache"
)

// BenchmarkLRUSet benchmarks LRU cache set operations
func BenchmarkLRUSet(b *testing.B) {
	lru, err := cache.NewLRUCache(1000)
	if err != nil {
		b.Fatalf("Failed to create LRU cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Set("key", "value")
	}
}

// BenchmarkLRUGet benchmarks LRU cache get operations
func BenchmarkLRUGet(b *testing.B) {
	lru, err := cache.NewLRUCache(1000)
	if err != nil {
		b.Fatalf("Failed to create LRU cache: %v", err)
	}
	lru.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Get("key")
	}
}

// BenchmarkLRUSetGet benchmarks combined set and get operations
func BenchmarkLRUSetGet(b *testing.B) {
	lru, err := cache.NewLRUCache(1000)
	if err != nil {
		b.Fatalf("Failed to create LRU cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i % 100))
		lru.Set(key, i)
		lru.Get(key)
	}
}

// BenchmarkRedisSet benchmarks Redis set operations
func BenchmarkRedisSet(b *testing.B) {
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		b.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redisClient.Set("bench_key", "value", 10*time.Minute)
	}
}

// BenchmarkRedisGet benchmarks Redis get operations
func BenchmarkRedisGet(b *testing.B) {
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		b.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	redisClient.Set("bench_key", "value", 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redisClient.Get("bench_key")
	}
}

// BenchmarkRedisSetJSON benchmarks Redis JSON set operations
func BenchmarkRedisSetJSON(b *testing.B) {
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		b.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	data := map[string]interface{}{
		"id":   123,
		"name": "test",
		"tags": []string{"a", "b", "c"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redisClient.SetJSON("bench_json", data, 10*time.Minute)
	}
}

// BenchmarkRedisGetJSON benchmarks Redis JSON get operations
func BenchmarkRedisGetJSON(b *testing.B) {
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		b.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	data := map[string]interface{}{
		"id":   123,
		"name": "test",
	}
	redisClient.SetJSON("bench_json", data, 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		redisClient.GetJSON("bench_json", &result)
	}
}

// BenchmarkRedisPipeline benchmarks Redis pipeline operations
func BenchmarkRedisPipeline(b *testing.B) {
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		b.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pairs := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		redisClient.MSet(pairs)
		redisClient.MGet("key1", "key2", "key3")
	}
}
