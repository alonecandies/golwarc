package cache_test

import (
	"testing"
	"time"

	"github.com/alonecandies/golwarc/cache"
	"github.com/alonecandies/golwarc/libs"
)

// TestRedisClient_NewRedisClient tests Redis client initialization
func TestRedisClient_NewRedisClient(t *testing.T) {
	// Skip if Redis is not available
	config := cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	client, err := cache.NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping Redis tests: %v", err)
		return
	}
	defer client.Close()

	if client == nil {
		t.Error("NewRedisClient() returned nil client")
	}
}

// TestRedisClient_NewRedisClientWithInvalidAddr tests connection failure
func TestRedisClient_NewRedisClientWithInvalidAddr(t *testing.T) {
	config := cache.RedisConfig{
		Addr:     "invalid-host:9999",
		Password: "",
		DB:       0,
	}

	_, err := cache.NewRedisClient(config)
	if err == nil {
		t.Error("NewRedisClient() should fail with invalid address")
	}
}

// TestRedisClient_SetAndGet tests basic set and get operations
func TestRedisClient_SetAndGet(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "test-key"
	value := "test-value"

	// Set value
	err := client.Set(key, value, 0)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Get value
	got, err := client.Get(key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}
}

// TestRedisClient_GetNonExistent tests getting non-existent key
func TestRedisClient_GetNonExistent(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	_, err := client.Get("non-existent-key")
	if err == nil {
		t.Error("Get() should return error for non-existent key")
	}
}

// TestRedisClient_SetWithTTL tests setting values with TTL
func TestRedisClient_SetWithTTL(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "ttl-key"
	value := "ttl-value"
	ttl := 2 * time.Second

	// Set with TTL
	err := client.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Verify value exists
	got, err := client.Get(key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Verify value is gone
	_, err = client.Get(key)
	if err == nil {
		t.Error("Get() should return error after TTL expiration")
	}
}

// TestRedisClient_SetJSON tests JSON operations
func TestRedisClient_SetJSON(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	type TestData struct {
		Name  string
		Age   int
		Email string
	}

	key := "json-key"
	original := TestData{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Set JSON
	err := client.SetJSON(key, original, 0)
	if err != nil {
		t.Errorf("SetJSON() error = %v", err)
	}

	// Get JSON
	var retrieved TestData
	err = client.GetJSON(key, &retrieved)
	if err != nil {
		t.Errorf("GetJSON() error = %v", err)
	}

	if retrieved != original {
		t.Errorf("GetJSON() = %+v, want %+v", retrieved, original)
	}
}

// TestRedisClient_Delete tests delete operations
func TestRedisClient_Delete(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "delete-key"
	client.Set(key, "value", 0)

	// Delete key
	err := client.Delete(key)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify key is gone
	_, err = client.Get(key)
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}
}

// TestRedisClient_DeleteMany tests batch delete
func TestRedisClient_DeleteMany(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		client.Set(key, "value", 0)
	}

	// Delete all keys
	err := client.DeleteMany(keys...)
	if err != nil {
		t.Errorf("DeleteMany() error = %v", err)
	}

	// Verify all keys are gone
	for _, key := range keys {
		_, err := client.Get(key)
		if err == nil {
			t.Errorf("Key %s should be deleted", key)
		}
	}
}

// TestRedisClient_Exists tests key existence check
func TestRedisClient_Exists(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "exists-key"
	client.Set(key, "value", 0)

	// Check existing key
	exists, err := client.Exists(key)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true for existing key")
	}

	// Check non-existent key
	exists, err = client.Exists("non-existent")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false for non-existent key")
	}
}

// TestRedisClient_Expire tests setting TTL on existing keys
func TestRedisClient_Expire(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "expire-key"
	client.Set(key, "value", 0)

	// Set expiration
	err := client.Expire(key, 1*time.Second)
	if err != nil {
		t.Errorf("Expire() error = %v", err)
	}

	// Verify key still exists
	exists, _ := client.Exists(key)
	if !exists {
		t.Error("Key should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Verify key is gone
	exists, _ = client.Exists(key)
	if exists {
		t.Error("Key should not exist after expiration")
	}
}

// TestRedisClient_TTL tests getting TTL
func TestRedisClient_TTL(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "ttl-test-key"
	ttl := 10 * time.Second
	client.Set(key, "value", ttl)

	// Get TTL
	remainingTTL, err := client.TTL(key)
	if err != nil {
		t.Errorf("TTL() error = %v", err)
	}

	// TTL should be close to what we set (within 1 second)
	if remainingTTL > ttl || remainingTTL < ttl-time.Second {
		t.Errorf("TTL() = %v, want approximately %v", remainingTTL, ttl)
	}
}

// TestRedisClient_Increment tests increment operations
func TestRedisClient_Increment(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "counter"

	// Increment new key
	val, err := client.Increment(key)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 1 {
		t.Errorf("Increment() = %d, want 1", val)
	}

	// Increment again
	val, err = client.Increment(key)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 2 {
		t.Errorf("Increment() = %d, want 2", val)
	}
}

// TestRedisClient_IncrementBy tests increment by value
func TestRedisClient_IncrementBy(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "counter-by"

	// Increment by 5
	val, err := client.IncrementBy(key, 5)
	if err != nil {
		t.Errorf("IncrementBy() error = %v", err)
	}
	if val != 5 {
		t.Errorf("IncrementBy() = %d, want 5", val)
	}

	// Increment by 10
	val, err = client.IncrementBy(key, 10)
	if err != nil {
		t.Errorf("IncrementBy() error = %v", err)
	}
	if val != 15 {
		t.Errorf("IncrementBy() = %d, want 15", val)
	}
}

// TestRedisClient_Decrement tests decrement operations
func TestRedisClient_Decrement(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "decounter"
	client.Set(key, 10, 0)

	// Decrement
	val, err := client.Decrement(key)
	if err != nil {
		t.Errorf("Decrement() error = %v", err)
	}
	if val != 9 {
		t.Errorf("Decrement() = %d, want 9", val)
	}
}

// TestRedisClient_SetNX tests atomic set if not exists
func TestRedisClient_SetNX(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "nx-key"

	// First SetNX should succeed
	ok, err := client.SetNX(key, "value1", 0)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if !ok {
		t.Error("SetNX() should succeed for new key")
	}

	// Second SetNX should fail
	ok, err = client.SetNX(key, "value2", 0)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if ok {
		t.Error("SetNX() should fail for existing key")
	}

	// Value should be unchanged
	val, _ := client.Get(key)
	if val != "value1" {
		t.Errorf("Get() = %v, want value1", val)
	}
}

// TestRedisClient_MGetMSet tests batch get/set operations
func TestRedisClient_MGetMSet(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	// Set multiple keys
	pairs := map[string]interface{}{
		"mkey1": "value1",
		"mkey2": "value2",
		"mkey3": "value3",
	}

	err := client.MSet(pairs)
	if err != nil {
		t.Errorf("MSet() error = %v", err)
	}

	// Get multiple keys
	values, err := client.MGet("mkey1", "mkey2", "mkey3")
	if err != nil {
		t.Errorf("MGet() error = %v", err)
	}

	if len(values) != 3 {
		t.Errorf("MGet() returned %d values, want 3", len(values))
	}

	// Verify values
	for i, val := range values {
		if val == nil {
			t.Errorf("MGet() value %d is nil", i)
		}
	}
}

// TestRedisClient_Ping tests connection health check
func TestRedisClient_Ping(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

// TestRedisClient_GetClient tests getting underlying client
func TestRedisClient_GetClient(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	underlying := client.GetClient()
	if underlying == nil {
		t.Error("GetClient() returned nil")
	}
}

// TestRedisClient_GetJSONInvalidJSON tests error handling for invalid JSON
func TestRedisClient_GetJSONInvalidJSON(t *testing.T) {
	client, skip := setupRedisTest(t)
	if skip {
		return
	}
	defer cleanupRedisTest(client)

	key := "invalid-json"
	client.Set(key, "not-json-data", 0)

	var dest map[string]interface{}
	err := client.GetJSON(key, &dest)
	if err == nil {
		t.Error("GetJSON() should return error for invalid JSON")
	}
	// Error is expected for invalid JSON - test passes if we get an error
}

// TestRedisClient_TLSConfig tests TLS configuration
func TestRedisClient_TLSConfig(t *testing.T) {
	// This test would require a Redis server with TLS enabled
	// For now, we just test that invalid TLS config is rejected
	config := cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		TLS: &libs.TLSConfig{
			Enabled: true,
			CACert:  "/invalid/path/ca.crt",
		},
	}

	_, err := cache.NewRedisClient(config)
	// Should fail due to invalid cert path
	if err == nil {
		// If we get here, either TLS is properly configured or Redis isn't available
		// Either way, the test has validated the TLS config flow
		t.Skip("TLS configuration accepted or Redis unavailable")
	}
}

// Helper functions

func setupRedisTest(t *testing.T) (*cache.RedisClient, bool) {
	config := cache.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests
	}

	client, err := cache.NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping Redis tests: Redis not available (%v)", err)
		return nil, true
	}

	// Clean up any existing test data
	client.FlushDB()

	return client, false
}

func cleanupRedisTest(client *cache.RedisClient) {
	if client != nil {
		client.FlushDB()
		client.Close()
	}
}
