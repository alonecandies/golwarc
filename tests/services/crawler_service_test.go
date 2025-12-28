package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alonecandies/golwarc/mocks"
	"github.com/alonecandies/golwarc/models"
	"github.com/alonecandies/golwarc/services"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// =============================================================================
// CrawlerService Unit Tests
// =============================================================================

func TestNewCrawlerService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockCache := &mocks.MockCacheClient{}
	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)

	if service == nil {
		t.Fatal("Expected non-nil CrawlerService")
	}
}

func TestCrawlerService_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		migrateErr  error
		expectError bool
	}{
		{
			name:        "success",
			migrateErr:  nil,
			expectError: false,
		},
		{
			name:        "migration error",
			migrateErr:  errors.New("migration failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			mockCache := &mocks.MockCacheClient{}
			mockDB := &mocks.MockDatabaseClient{
				MigrateFunc: func(models ...interface{}) error {
					return tt.migrateErr
				},
			}

			service := services.NewCrawlerService(logger, mockCache, mockDB)
			err := service.Initialize()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCrawlerService_CrawlAndStore_CacheHit(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Mock cache that returns a hit
	mockCache := &mocks.MockCacheClient{
		ExistsFunc: func(key string) (bool, error) {
			return true, nil // Cache hit
		},
	}

	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)
	err := service.CrawlAndStore("https://example.com")

	if err != nil {
		t.Errorf("Expected no error on cache hit, got: %v", err)
	}
}

func TestCrawlerService_CrawlAndStore_NilCache(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := &mocks.MockDatabaseClient{}

	// Service with nil cache - should still work (just skip caching)
	service := services.NewCrawlerService(logger, nil, mockDB)

	// This will try to crawl, which may fail due to network
	// The important thing is it doesn't panic with nil cache
	_ = service.CrawlAndStore("https://invalid-test-url-12345.invalid")
}

func TestCrawlerService_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock SQL database for GORM
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	tests := []struct {
		name          string
		setupMock     func()
		expectedCount int64
		expectError   bool
	}{
		{
			name: "success with pages",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `pages`").WillReturnRows(rows)
			},
			expectedCount: 42,
			expectError:   false,
		},
		{
			name: "success with zero pages",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `pages`").WillReturnRows(rows)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "database error",
			setupMock: func() {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `pages`").WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			mockCache := &mocks.MockCacheClient{}
			mockDB := &mocks.MockDatabaseClient{
				DB: gormDB,
			}

			service := services.NewCrawlerService(logger, mockCache, mockDB)
			stats, err := service.GetStats()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if stats == nil {
				t.Fatal("Expected non-nil stats")
			}

			pageCount, ok := stats["total_pages_crawled"].(int64)
			if !ok {
				t.Fatal("total_pages_crawled not found or wrong type")
			}

			if pageCount != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, pageCount)
			}

			// Verify cache_enabled field
			cacheEnabled, ok := stats["cache_enabled"].(bool)
			if !ok {
				t.Error("cache_enabled not found or wrong type")
			}
			if !cacheEnabled {
				t.Error("Expected cache_enabled to be true")
			}

			// Verify database_connected field
			dbConnected, ok := stats["database_connected"].(bool)
			if !ok {
				t.Error("database_connected not found or wrong type")
			}
			if !dbConnected {
				t.Error("Expected database_connected to be true")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unmet expectations: %v", err)
			}
		})
	}
}

func TestCrawlerService_GetStats_NilCache(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock SQL database for GORM
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(10)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `pages`").WillReturnRows(rows)

	mockDB := &mocks.MockDatabaseClient{
		DB: gormDB,
	}

	service := services.NewCrawlerService(logger, nil, mockDB)
	stats, err := service.GetStats()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With nil cache, cache_enabled should be false
	cacheEnabled, _ := stats["cache_enabled"].(bool)
	if cacheEnabled {
		t.Error("Expected cache_enabled to be false with nil cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestCrawlerService_GetRecentPages(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock SQL database for GORM
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	tests := []struct {
		name         string
		limit        int
		setupMock    func()
		expectedRows int
		expectError  bool
	}{
		{
			name:  "success with multiple pages",
			limit: 5,
			setupMock: func() {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "url", "title", "domain", "status", "html", "created_at", "updated_at"}).
					AddRow(1, "https://example.com", "Example", "example.com", 200, "<html></html>", now, now).
					AddRow(2, "https://test.com", "Test", "test.com", 200, "<html></html>", now, now)
				// Match SQL with regex to handle parameter placeholders
				mock.ExpectQuery("SELECT \\* FROM `pages` WHERE `pages`.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT").WillReturnRows(rows)
			},
			expectedRows: 2,
			expectError:  false,
		},
		{
			name:  "success with zero pages",
			limit: 10,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "url", "title", "domain", "status", "html", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM `pages` WHERE `pages`.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT").WillReturnRows(rows)
			},
			expectedRows: 0,
			expectError:  false,
		},
		{
			name:  "limit 1",
			limit: 1,
			setupMock: func() {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "url", "title", "domain", "status", "html", "created_at", "updated_at"}).
					AddRow(1, "https://example.com", "Example", "example.com", 200, "<html></html>", now, now)
				mock.ExpectQuery("SELECT \\* FROM `pages` WHERE `pages`.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT").WillReturnRows(rows)
			},
			expectedRows: 1,
			expectError:  false,
		},
		{
			name:  "database error",
			limit: 5,
			setupMock: func() {
				mock.ExpectQuery("SELECT \\* FROM `pages` WHERE `pages`.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT").WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			mockCache := &mocks.MockCacheClient{}
			mockDB := &mocks.MockDatabaseClient{
				DB: gormDB,
			}

			service := services.NewCrawlerService(logger, mockCache, mockDB)
			pages, err := service.GetRecentPages(tt.limit)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(pages) != tt.expectedRows {
				t.Errorf("Expected %d pages, got %d", tt.expectedRows, len(pages))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unmet expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// Integration-like Tests with Mocks
// =============================================================================

func TestCrawlerService_InitializeWithModels(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockCache := &mocks.MockCacheClient{}

	var migratedModels []interface{}
	mockDB := &mocks.MockDatabaseClient{
		MigrateFunc: func(models ...interface{}) error {
			migratedModels = models
			return nil
		},
	}

	service := services.NewCrawlerService(logger, mockCache, mockDB)
	err := service.Initialize()

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify that 3 models were migrated (Page, Product, Article)
	if len(migratedModels) != 3 {
		t.Errorf("Expected 3 models to be migrated, got %d", len(migratedModels))
	}

	// Verify the types
	_, isPage := migratedModels[0].(*models.Page)
	_, isProduct := migratedModels[1].(*models.Product)
	_, isArticle := migratedModels[2].(*models.Article)

	if !isPage || !isProduct || !isArticle {
		t.Error("Migrated models don't match expected types")
	}
}

func TestCrawlerService_CacheKeyFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)

	var checkedKey string
	mockCache := &mocks.MockCacheClient{
		ExistsFunc: func(key string) (bool, error) {
			checkedKey = key
			return true, nil // Return hit to avoid actual crawl
		},
	}
	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)
	_ = service.CrawlAndStore("https://example.com")

	expectedKey := "page:https://example.com"
	if checkedKey != expectedKey {
		t.Errorf("Expected cache key %q, got %q", expectedKey, checkedKey)
	}
}

func TestCrawlerService_CacheKeyFormat_DifferentURLs(t *testing.T) {
	tests := []struct {
		url         string
		expectedKey string
	}{
		{"https://example.com", "page:https://example.com"},
		{"https://example.com/path", "page:https://example.com/path"},
		{"http://localhost:8080", "page:http://localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			var checkedKey string
			mockCache := &mocks.MockCacheClient{
				ExistsFunc: func(key string) (bool, error) {
					checkedKey = key
					return true, nil
				},
			}
			mockDB := &mocks.MockDatabaseClient{}

			service := services.NewCrawlerService(logger, mockCache, mockDB)
			_ = service.CrawlAndStore(tt.url)

			if checkedKey != tt.expectedKey {
				t.Errorf("Expected cache key %q, got %q", tt.expectedKey, checkedKey)
			}
		})
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCrawlerService_LoggerUsage(t *testing.T) {
	// Test with production-like logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockCache := &mocks.MockCacheClient{
		ExistsFunc: func(key string) (bool, error) {
			return true, nil
		},
	}
	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)
	err := service.CrawlAndStore("https://example.com")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCrawlerService_CacheExistsError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Mock cache that returns an error on Exists
	mockCache := &mocks.MockCacheClient{
		ExistsFunc: func(key string) (bool, error) {
			return false, errors.New("cache connection error")
		},
	}
	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)

	// Should proceed to crawl despite cache error
	// Will fail on crawl, but shouldn't panic
	_ = service.CrawlAndStore("https://invalid-test-url.invalid")
}

func TestCrawlerService_GetRecentPages_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock SQL database for GORM
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	tests := []struct {
		name  string
		limit int
	}{
		{"very large limit", 10000},
		{"negative limit", -1}, // GORM may handle this gracefully
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock expects a query regardless of limit
			rows := sqlmock.NewRows([]string{"id", "url", "title", "domain", "status", "html", "created_at", "updated_at"})
			mock.ExpectQuery("SELECT \\* FROM `pages`").WillReturnRows(rows)

			mockCache := &mocks.MockCacheClient{}
			mockDB := &mocks.MockDatabaseClient{
				DB: gormDB,
			}

			service := services.NewCrawlerService(logger, mockCache, mockDB)
			_, _ = service.GetRecentPages(tt.limit)
		})
	}
}

// =============================================================================
// Mock Verification Tests
// =============================================================================

func TestMockDatabaseClient_ImplementsInterface(t *testing.T) {
	mock := &mocks.MockDatabaseClient{}

	// Verify all interface methods are callable
	_ = mock.Create(&models.Page{})
	_ = mock.Find(&[]models.Page{})
	_ = mock.First(&models.Page{})
	_ = mock.Update(&models.Page{}, "title", "new")
	_ = mock.Updates(&models.Page{}, map[string]interface{}{})
	_ = mock.Delete(&models.Page{})
	_ = mock.Migrate(&models.Page{})
	_ = mock.Ping()
	_ = mock.Close()
	_ = mock.Transaction(func(tx *gorm.DB) error { return nil })

	t.Log("MockDatabaseClient implements all required methods")
}

func TestMockCacheClient_ImplementsInterface(t *testing.T) {
	mock := &mocks.MockCacheClient{}

	// Verify all interface methods are callable
	_, _ = mock.Get("key")
	_ = mock.Set("key", "value", 0)
	_ = mock.Delete("key")
	_, _ = mock.Exists("key")
	_ = mock.Close()
	_ = mock.Ping()
	_ = mock.GetJSON("key", &struct{}{})
	_ = mock.SetJSON("key", struct{}{}, 0)

	t.Log("MockCacheClient implements all required methods")
}

// =============================================================================
// Constructor Edge Cases
// =============================================================================

func TestNewCrawlerService_AllNilDependencies(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// All nil except logger
	service := services.NewCrawlerService(logger, nil, nil)
	if service == nil {
		t.Fatal("Expected non-nil service even with nil dependencies")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCrawlerService_CacheHit(b *testing.B) {
	logger := zap.NewNop()
	mockCache := &mocks.MockCacheClient{
		ExistsFunc: func(key string) (bool, error) {
			return true, nil
		},
	}
	mockDB := &mocks.MockDatabaseClient{}

	service := services.NewCrawlerService(logger, mockCache, mockDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.CrawlAndStore("https://example.com")
	}
}
