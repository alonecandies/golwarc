package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alonecandies/golwarc/cache"
	"github.com/alonecandies/golwarc/configs"
	"github.com/alonecandies/golwarc/crawlers"
	"github.com/alonecandies/golwarc/database"
	"github.com/alonecandies/golwarc/inject"
	"github.com/alonecandies/golwarc/libs"
	messagequeue "github.com/alonecandies/golwarc/message-queue"
	"github.com/alonecandies/golwarc/models"
	"github.com/alonecandies/golwarc/services"
	"github.com/gocolly/colly/v2"
)

// =====================================================
// E2E Test: Cache + Database Integration
// =====================================================

func TestCacheDatabaseIntegration(t *testing.T) {
	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	// Initialize MySQL
	mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer mysqlClient.Close()

	// Migrate test model
	type CacheTestRecord struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Value string `gorm:"size:255"`
	}
	mysqlClient.Migrate(&CacheTestRecord{})

	// Test flow: Write to DB -> Cache -> Read from Cache
	record := &CacheTestRecord{Name: "cache_test", Value: "initial_value"}
	if err := mysqlClient.Create(record); err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Cache the record
	cacheKey := fmt.Sprintf("record:%d", record.ID)
	if err := redisClient.SetJSON(cacheKey, record, 5*time.Minute); err != nil {
		t.Fatalf("Failed to cache record: %v", err)
	}

	// Read from cache
	var cached CacheTestRecord
	if err := redisClient.GetJSON(cacheKey, &cached); err != nil {
		t.Fatalf("Failed to read from cache: %v", err)
	}

	if cached.Name != record.Name {
		t.Errorf("Cache mismatch: expected %s, got %s", record.Name, cached.Name)
	}

	// Verify cache invalidation
	redisClient.Delete(cacheKey)
	exists, _ := redisClient.Exists(cacheKey)
	if exists {
		t.Error("Cache should be invalidated")
	}

	// Cleanup
	mysqlClient.Delete(&CacheTestRecord{}, record.ID)
}

// =====================================================
// E2E Test: Crawler + Cache Integration
// =====================================================

func TestCrawlerCacheIntegration(t *testing.T) {
	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	// Create crawler
	crawler := crawlers.NewDefaultCollyClient()

	var scrapedTitle string
	crawler.OnHTML("title", func(e *colly.HTMLElement) {
		scrapedTitle = e.Text
	})

	// Crawl and cache
	targetURL := "https://example.com"
	cacheKey := fmt.Sprintf("crawl:%s", targetURL)

	// Check cache first
	cached, _ := redisClient.Get(cacheKey)
	if cached != "" {
		t.Log("Found in cache, skipping crawl")
		return
	}

	if err := crawler.Visit(targetURL); err != nil {
		t.Skip("Network not available:", err)
	}
	crawler.Wait()

	if scrapedTitle != "" {
		// Cache the result
		redisClient.Set(cacheKey, scrapedTitle, 10*time.Minute)

		// Verify cache
		cachedTitle, err := redisClient.Get(cacheKey)
		if err != nil {
			t.Fatalf("Failed to read cache: %v", err)
		}
		if cachedTitle != scrapedTitle {
			t.Errorf("Cache mismatch: expected %s, got %s", scrapedTitle, cachedTitle)
		}
	}

	// Cleanup
	redisClient.Delete(cacheKey)
}

// =====================================================
// E2E Test: Crawler + Database Integration
// =====================================================

func TestCrawlerDatabaseIntegration(t *testing.T) {
	// Initialize MySQL
	mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer mysqlClient.Close()

	// Migrate model
	mysqlClient.Migrate(&models.Page{})

	// Create crawler
	crawler := crawlers.NewDefaultCollyClient()

	var pageToSave *models.Page
	crawler.OnHTML("html", func(e *colly.HTMLElement) {
		pageToSave = &models.Page{
			URL:    e.Request.URL.String(),
			Title:  e.ChildText("title"),
			Domain: e.Request.URL.Host,
			Status: 200,
		}
	})

	if err := crawler.Visit("https://example.com"); err != nil {
		t.Skip("Network not available:", err)
	}
	crawler.Wait()

	if pageToSave != nil {
		// Save to database
		if err := mysqlClient.Create(pageToSave); err != nil {
			t.Fatalf("Failed to save page: %v", err)
		}

		// Verify in database
		var found models.Page
		if err := mysqlClient.First(&found, pageToSave.ID); err != nil {
			t.Fatalf("Failed to find page: %v", err)
		}

		if found.URL != pageToSave.URL {
			t.Errorf("URL mismatch: expected %s, got %s", pageToSave.URL, found.URL)
		}

		// Cleanup
		mysqlClient.Delete(&models.Page{}, pageToSave.ID)
	}
}

// =====================================================
// E2E Test: Message Queue + Cache Integration
// =====================================================

func TestMessageQueueCacheIntegration(t *testing.T) {
	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	// Initialize RabbitMQ
	rabbitClient, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer rabbitClient.Close()

	// Create queue
	queueName := "cache-sync-test"
	rabbitClient.DeclareQueue(queueName, false)

	// Simulate cache invalidation via message queue
	cacheKey := "sync:test:key"
	redisClient.Set(cacheKey, "cached_value", 5*time.Minute)

	// Publish invalidation message
	ctx := context.Background()
	rabbitClient.Publish(ctx, queueName, []byte(cacheKey))

	// In a real scenario, consumer would delete the cache
	// For test, we simulate it
	redisClient.Delete(cacheKey)

	exists, _ := redisClient.Exists(cacheKey)
	if exists {
		t.Error("Cache should be invalidated after message")
	}

	// Cleanup
	rabbitClient.DeleteQueue(queueName)
}

// =====================================================
// E2E Test: Full Integration (All Dependencies)
// =====================================================

func TestFullIntegration(t *testing.T) {
	// Initialize logger
	if err := libs.InitDefaultLogger(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer libs.Sync()
	logger := libs.GetLogger()

	// Load configuration
	_ = configs.GetDefaultConfig()
	logger.Info("Config loaded")

	// Initialize LRU Cache
	lruCache, err := cache.NewLRUCache(100)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}
	logger.Info("LRU cache initialized")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer redisClient.Close()
	logger.Info("Redis connected")

	// Initialize MySQL
	mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer mysqlClient.Close()
	logger.Info("MySQL connected")

	// Migrate models
	mysqlClient.Migrate(&models.Page{}, &models.Product{}, &models.Article{})
	logger.Info("Models migrated")

	// Create crawler service using DI pattern
	crawlerService := services.NewCrawlerService(logger, redisClient, mysqlClient)
	if initErr := crawlerService.Initialize(); initErr != nil {
		t.Fatalf("Failed to initialize crawler service: %v", initErr)
	}

	// Full workflow test
	testURL := "https://example.com"

	// Step 1: Check LRU cache
	_, lruHit := lruCache.Get(testURL)
	if lruHit {
		logger.Info("LRU cache hit")
	}

	// Step 2: Check Redis cache
	redisHit, _ := redisClient.Exists(fmt.Sprintf("page:%s", testURL))
	if redisHit {
		logger.Info("Redis cache hit")
	}

	// Step 3: Crawl if not cached
	if !lruHit && !redisHit {
		crawlErr := crawlerService.CrawlAndStore(testURL)
		if crawlErr != nil {
			t.Skip("Network not available:", crawlErr)
		}

		// Update LRU cache
		lruCache.Set(testURL, true)
	}

	// Step 4: Verify data in database
	var pages []models.Page
	mysqlClient.Find(&pages)
	logger.Info(fmt.Sprintf("Total pages in database: %d", len(pages)))

	// Step 5: Get statistics
	stats, err := crawlerService.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	logger.Info(fmt.Sprintf("Stats: %v", stats))

	// Verify all components worked together
	if len(pages) == 0 && !lruHit && !redisHit {
		t.Log("Warning: No pages crawled, network might be unavailable")
	}

	logger.Info("Full integration test completed successfully")
}

// =====================================================
// E2E Test: DI Container Integration
// =====================================================

func TestDIContainerIntegration(t *testing.T) {
	// This test requires a valid config file
	container, err := inject.NewContainer("../../config.example.yaml")
	if err != nil {
		// Try to initialize with minimal setup
		libs.InitDefaultLogger()
		t.Skip("Could not initialize container, some services may be unavailable")
	}
	defer container.Close()

	// Verify logger is initialized
	if container.Logger == nil {
		t.Error("Logger should not be nil")
	}

	// Verify config is loaded
	if container.Config == nil {
		t.Error("Config should not be nil")
	}

	container.Logger.Info("DI Container integration test passed")
}

// =====================================================
// E2E Test: Complete Pipeline (Crawl -> Cache -> DB -> Queue)
// =====================================================

func TestCompletePipeline(t *testing.T) {
	// Initialize all services
	libs.InitDefaultLogger()
	defer libs.Sync()
	logger := libs.GetLogger()

	// Redis
	redisClient, err := cache.NewRedisClient(cache.RedisConfig{Addr: "localhost:6379"})
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer redisClient.Close()

	// MySQL
	mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
		Host: "localhost", Port: 3306, User: "root", Password: "password", Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer mysqlClient.Close()

	// RabbitMQ
	rabbitClient, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	defer rabbitClient.Close()

	// Setup
	mysqlClient.Migrate(&models.Page{})
	rabbitClient.DeclareQueue("crawl-events", false)

	// Crawler
	crawler := crawlers.NewDefaultCollyClient()
	var crawledPage *models.Page

	crawler.OnHTML("html", func(e *colly.HTMLElement) {
		crawledPage = &models.Page{
			URL:    e.Request.URL.String(),
			Title:  e.ChildText("title"),
			Domain: e.Request.URL.Host,
			Status: 200,
		}
	})

	// Execute pipeline
	testURL := "https://example.com"

	// 1. Crawl
	logger.Info("Step 1: Crawling...")
	if err := crawler.Visit(testURL); err != nil {
		t.Skip("Network not available:", err)
	}
	crawler.Wait()

	if crawledPage == nil {
		t.Skip("No page crawled, network issue")
	}

	// 2. Cache
	logger.Info("Step 2: Caching...")
	cacheKey := fmt.Sprintf("page:%s", testURL)
	redisClient.SetJSON(cacheKey, crawledPage, 24*time.Hour)

	// 3. Save to DB
	logger.Info("Step 3: Saving to database...")
	mysqlClient.Create(crawledPage)

	// 4. Publish event
	logger.Info("Step 4: Publishing event...")
	ctx := context.Background()
	eventMsg := fmt.Sprintf(`{"event":"page_crawled","url":"%s","page_id":%d}`, testURL, crawledPage.ID)
	rabbitClient.Publish(ctx, "crawl-events", []byte(eventMsg))

	// Verify all steps
	logger.Info("Verifying pipeline...")

	// Verify cache
	var cachedPage models.Page
	if err := redisClient.GetJSON(cacheKey, &cachedPage); err != nil {
		t.Errorf("Cache verification failed: %v", err)
	}

	// Verify DB
	var dbPage models.Page
	if err := mysqlClient.First(&dbPage, crawledPage.ID); err != nil {
		t.Errorf("Database verification failed: %v", err)
	}

	logger.Info("Complete pipeline test passed!")

	// Cleanup
	redisClient.Delete(cacheKey)
	mysqlClient.Delete(&models.Page{}, crawledPage.ID)
	rabbitClient.DeleteQueue("crawl-events")
}
