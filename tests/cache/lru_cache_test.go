package cache_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alonecandies/golwarc/cache"
)

func TestNewLRUCache(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{
			name:    "valid size",
			size:    100,
			wantErr: false,
		},
		{
			name:    "zero size",
			size:    0,
			wantErr: true,
		},
		{
			name:    "negative size",
			size:    -1,
			wantErr: true,
		},
		{
			name:    "size of 1",
			size:    1,
			wantErr: false,
		},
		{
			name:    "large size",
			size:    10000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lruCache, err := cache.NewLRUCache(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLRUCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && lruCache == nil {
				t.Error("NewLRUCache() returned nil cache")
			}
			if !tt.wantErr && lruCache.Len() != 0 {
				t.Errorf("NewLRUCache() initial length = %d, want 0", lruCache.Len())
			}
		})
	}
}

func TestLRUCache_SetAndGet(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test basic set and get
	key := "test-key"
	value := "test-value"

	// Set value - returns true if key is new, false if evicted
	lruCache.Set(key, value)

	got, found := lruCache.Get(key)
	if !found {
		t.Error("Get() should find the key")
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}
}

func TestLRUCache_SetUpdate(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set initial value
	lruCache.Set("key", "value1")

	// Update value
	lruCache.Set("key", "value2")

	// Get should return updated value
	val, found := lruCache.Get("key")
	if !found {
		t.Error("Get() should find the key")
	}
	if val != "value2" {
		t.Errorf("Get() = %v, want value2", val)
	}

	// Length should still be 1
	if lruCache.Len() != 1 {
		t.Errorf("Len() = %d, want 1", lruCache.Len())
	}
}

func TestLRUCache_GetNonExistent(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	_, found := lruCache.Get("non-existent")
	if found {
		t.Error("Get() should not find non-existent key")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := "test-key"
	lruCache.Set(key, "value")

	// Delete existing key
	deleted := lruCache.Delete(key)
	if !deleted {
		t.Error("Delete() should return true for existing key")
	}

	// Verify key is gone
	_, found := lruCache.Get(key)
	if found {
		t.Error("Get() should not find deleted key")
	}

	// Length should be 0
	if lruCache.Len() != 0 {
		t.Errorf("Len() after Delete() = %d, want 0", lruCache.Len())
	}

	// Delete non-existent key
	deleted = lruCache.Delete("non-existent")
	if deleted {
		t.Error("Delete() should return false for non-existent key")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add multiple items
	lruCache.Set("key1", "value1")
	lruCache.Set("key2", "value2")
	lruCache.Set("key3", "value3")

	if lruCache.Len() != 3 {
		t.Errorf("Len() = %d, want 3", lruCache.Len())
	}

	// Clear cache
	lruCache.Clear()

	if lruCache.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", lruCache.Len())
	}

	// Verify items are gone
	_, found := lruCache.Get("key1")
	if found {
		t.Error("Get() should not find cleared key")
	}

	// Should be able to add items after clear
	lruCache.Set("new-key", "new-value")
	if lruCache.Len() != 1 {
		t.Errorf("Len() after adding to cleared cache = %d, want 1", lruCache.Len())
	}
}

func TestLRUCache_Contains(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := "test-key"
	lruCache.Set(key, "value")

	if !lruCache.Contains(key) {
		t.Error("Contains() should return true for existing key")
	}

	if lruCache.Contains("non-existent") {
		t.Error("Contains() should return false for non-existent key")
	}

	// Contains should not affect LRU order
	lruCache.Set("key2", "value2")
	lruCache.Contains(key) // Check first key
	lruCache.Set("key3", "value3")
	// First key should still be there because Contains doesn't update LRU
	if !lruCache.Contains(key) {
		t.Error("Contains() should not affect LRU order")
	}
}

func TestLRUCache_Peek(t *testing.T) {
	lruCache, err := cache.NewLRUCache(2)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add items
	lruCache.Set("key1", "value1")
	lruCache.Set("key2", "value2")

	// Peek should not affect LRU order
	val, found := lruCache.Peek("key1")
	if !found {
		t.Error("Peek() should find existing key")
	}
	if val != "value1" {
		t.Errorf("Peek() = %v, want value1", val)
	}

	// Peek on non-existent key
	_, found = lruCache.Peek("non-existent")
	if found {
		t.Error("Peek() should not find non-existent key")
	}

	// Add another item - key1 should be evicted if Peek affected LRU
	lruCache.Set("key3", "value3")

	// If Peek didn't affect LRU, key1 should still be there
	// (key2 was accessed by Set, so key1 is oldest)
	_, found = lruCache.Get("key2")
	if !found {
		t.Error("key2 should still exist")
	}
}

func TestLRUCache_Keys(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Empty cache should return empty slice
	keys := lruCache.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys() for empty cache returned %d keys, want 0", len(keys))
	}

	// Add items in order
	lruCache.Set("key1", "value1")
	lruCache.Set("key2", "value2")
	lruCache.Set("key3", "value3")

	keys = lruCache.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	// Check if all keys are present
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for _, k := range expectedKeys {
		if !keyMap[k] {
			t.Errorf("Keys() missing expected key %s", k)
		}
	}
}

func TestLRUCache_Resize(t *testing.T) {
	lruCache, err := cache.NewLRUCache(3)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Fill cache
	lruCache.Set("key1", "value1")
	lruCache.Set("key2", "value2")
	lruCache.Set("key3", "value3")

	// Resize to smaller size
	evicted, err := lruCache.Resize(2)
	if err != nil {
		t.Errorf("Resize() error = %v", err)
	}
	if evicted != 1 {
		t.Errorf("Resize() evicted %d items, want 1", evicted)
	}

	if lruCache.Len() != 2 {
		t.Errorf("Len() after Resize() = %d, want 2", lruCache.Len())
	}

	// Resize to larger size
	evicted, err = lruCache.Resize(10)
	if err != nil {
		t.Errorf("Resize() error = %v", err)
	}
	if evicted != 0 {
		t.Errorf("Resize() to larger size evicted %d items, want 0", evicted)
	}

	// Test invalid resize
	_, err = lruCache.Resize(0)
	if err == nil {
		t.Error("Resize(0) should return error")
	}

	_, err = lruCache.Resize(-1)
	if err == nil {
		t.Error("Resize(-1) should return error")
	}
}

func TestLRUCache_EvictionOrder(t *testing.T) {
	lruCache, err := cache.NewLRUCache(2)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add two items
	lruCache.Set("key1", "value1")
	lruCache.Set("key2", "value2")

	// Access key1 to make it recently used
	lruCache.Get("key1")

	// Add third item - key2 should be evicted
	lruCache.Set("key3", "value3")

	// key1 should still exist (was recently accessed)
	if !lruCache.Contains("key1") {
		t.Error("key1 should still exist (was recently used)")
	}

	// key2 should be evicted
	if lruCache.Contains("key2") {
		t.Error("key2 should be evicted (least recently used)")
	}

	// key3 should exist
	if !lruCache.Contains("key3") {
		t.Error("key3 should exist (just added)")
	}
}

func TestLRUCache_NilValue(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set nil value
	lruCache.Set("nil-key", nil)

	// Get nil value
	val, found := lruCache.Get("nil-key")
	if !found {
		t.Error("Get() should find key with nil value")
	}
	if val != nil {
		t.Errorf("Get() = %v, want nil", val)
	}

	// Contains should work with nil values
	if !lruCache.Contains("nil-key") {
		t.Error("Contains() should return true for key with nil value")
	}

	// Peek should work with nil values
	val, found = lruCache.Peek("nil-key")
	if !found || val != nil {
		t.Error("Peek() should find key with nil value")
	}
}

func TestLRUCache_DifferentTypes(t *testing.T) {
	lruCache, err := cache.NewLRUCache(10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test string
	lruCache.Set("string", "value")
	val, _ := lruCache.Get("string")
	if val.(string) != "value" {
		t.Error("Failed to store/retrieve string")
	}

	// Test int
	lruCache.Set("int", 123)
	val, _ = lruCache.Get("int")
	if val.(int) != 123 {
		t.Error("Failed to store/retrieve int")
	}

	// Test float
	lruCache.Set("float", 3.14)
	val, _ = lruCache.Get("float")
	if val.(float64) != 3.14 {
		t.Error("Failed to store/retrieve float")
	}

	// Test bool
	lruCache.Set("bool", true)
	val, _ = lruCache.Get("bool")
	if val.(bool) != true {
		t.Error("Failed to store/retrieve bool")
	}

	// Test struct
	type TestStruct struct {
		Name string
		Age  int
	}
	testObj := TestStruct{Name: "Test", Age: 30}
	lruCache.Set("struct", testObj)
	val, _ = lruCache.Get("struct")
	if val.(TestStruct).Name != "Test" {
		t.Error("Failed to store/retrieve struct")
	}

	// Test slice
	testSlice := []string{"a", "b", "c"}
	lruCache.Set("slice", testSlice)
	val, _ = lruCache.Get("slice")
	if len(val.([]string)) != 3 {
		t.Error("Failed to store/retrieve slice")
	}

	// Test map
	testMap := map[string]int{"one": 1, "two": 2}
	lruCache.Set("map", testMap)
	val, _ = lruCache.Get("map")
	if val.(map[string]int)["one"] != 1 {
		t.Error("Failed to store/retrieve map")
	}
}

func TestLRUCache_ConcurrentReadsWrites(t *testing.T) {
	lruCache, err := cache.NewLRUCache(100)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Note: LRU cache from hashicorp is thread-safe
	var wg sync.WaitGroup
	iterations := 1000

	// Multiple writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				lruCache.Set(key, j)
			}
		}(i)
	}

	// Multiple readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				lruCache.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestLRUCache_ConcurrentDeletes(t *testing.T) {
	lruCache, err := cache.NewLRUCache(100)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Pre-fill cache
	for i := 0; i < 50; i++ {
		lruCache.Set(fmt.Sprintf("key-%d", i), i)
	}

	var wg sync.WaitGroup

	// Multiple deleters
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(_ int) { // Explicitly ignore id to avoid unparam warning
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("key-%d", j)
				lruCache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// Cache should be empty or nearly empty
	if lruCache.Len() > 10 {
		t.Errorf("Len() after concurrent deletes = %d, expected <= 10", lruCache.Len())
	}
}

func TestLRUCache_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	lruCache, err := cache.NewLRUCache(1000)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	var wg sync.WaitGroup
	operations := 10000

	// Mixed operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(_ int) { // Explicitly ignore id to avoid unparam warning
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("key-%d", j%100)
				switch j % 5 {
				case 0:
					lruCache.Set(key, j)
				case 1:
					lruCache.Get(key)
				case 2:
					lruCache.Contains(key)
				case 3:
					lruCache.Peek(key)
				case 4:
					lruCache.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// No assertions - just ensure it doesn't crash
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLRUCache_Set(b *testing.B) {
	lruCache, _ := cache.NewLRUCache(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lruCache.Set(fmt.Sprintf("key-%d", i), i)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	lruCache, _ := cache.NewLRUCache(1000)
	for i := 0; i < 1000; i++ {
		lruCache.Set(fmt.Sprintf("key-%d", i), i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lruCache.Get(fmt.Sprintf("key-%d", i%1000))
	}
}

func BenchmarkLRUCache_SetGetMixed(b *testing.B) {
	lruCache, _ := cache.NewLRUCache(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			lruCache.Set(fmt.Sprintf("key-%d", i%1000), i)
		} else {
			lruCache.Get(fmt.Sprintf("key-%d", i%1000))
		}
	}
}

func BenchmarkLRUCache_ConcurrentAccess(b *testing.B) {
	lruCache, _ := cache.NewLRUCache(1000)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			if i%2 == 0 {
				lruCache.Set(key, i)
			} else {
				lruCache.Get(key)
			}
			i++
		}
	})
}
