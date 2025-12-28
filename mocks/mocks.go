package mocks

import (
	"time"

	"github.com/gocolly/colly/v2"
	"gorm.io/gorm"
)

// =============================================================================
// Mock Database Client
// =============================================================================

// MockDatabaseClient is a mock implementation of database.DatabaseClient
type MockDatabaseClient struct {
	DB           *gorm.DB
	CreateFunc   func(value interface{}) error
	FindFunc     func(dest interface{}, conds ...interface{}) error
	FirstFunc    func(dest interface{}, conds ...interface{}) error
	UpdateFunc   func(model interface{}, column string, value interface{}) error
	UpdatesFunc  func(model interface{}, values interface{}) error
	DeleteFunc   func(value interface{}, conds ...interface{}) error
	MigrateFunc  func(models ...interface{}) error
	PingFunc     func() error
	CloseFunc    func() error
	TransactFunc func(fn func(*gorm.DB) error) error
}

// GetDB returns the underlying GORM database instance
func (m *MockDatabaseClient) GetDB() *gorm.DB {
	return m.DB
}

// Create inserts a new record into the database
func (m *MockDatabaseClient) Create(value interface{}) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(value)
	}
	return nil
}

// Find retrieves records based on conditions
func (m *MockDatabaseClient) Find(dest interface{}, conds ...interface{}) error {
	if m.FindFunc != nil {
		return m.FindFunc(dest, conds...)
	}
	return nil
}

// First finds the first record ordered by primary key
func (m *MockDatabaseClient) First(dest interface{}, conds ...interface{}) error {
	if m.FirstFunc != nil {
		return m.FirstFunc(dest, conds...)
	}
	return nil
}

// Update updates a single column on a model
func (m *MockDatabaseClient) Update(model interface{}, column string, value interface{}) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(model, column, value)
	}
	return nil
}

// Updates updates multiple columns on a model
func (m *MockDatabaseClient) Updates(model interface{}, values interface{}) error {
	if m.UpdatesFunc != nil {
		return m.UpdatesFunc(model, values)
	}
	return nil
}

// Delete removes a record from the database
func (m *MockDatabaseClient) Delete(value interface{}, conds ...interface{}) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(value, conds...)
	}
	return nil
}

// Migrate automatically migrates the schema for the given models
func (m *MockDatabaseClient) Migrate(models ...interface{}) error {
	if m.MigrateFunc != nil {
		return m.MigrateFunc(models...)
	}
	return nil
}

// Ping checks the database connection
func (m *MockDatabaseClient) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

// Close closes the database connection
func (m *MockDatabaseClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Transaction executes a function within a database transaction
func (m *MockDatabaseClient) Transaction(fn func(*gorm.DB) error) error {
	if m.TransactFunc != nil {
		return m.TransactFunc(fn)
	}
	return nil
}

// =============================================================================
// Mock Cache Client
// =============================================================================

// MockCacheClient is a mock implementation of cache.JSONCacheClient
type MockCacheClient struct {
	Data        map[string]string
	GetFunc     func(key string) (string, error)
	SetFunc     func(key string, value interface{}, ttl time.Duration) error
	DeleteFunc  func(key string) error
	ExistsFunc  func(key string) (bool, error)
	CloseFunc   func() error
	PingFunc    func() error
	GetJSONFunc func(key string, dest interface{}) error
	SetJSONFunc func(key string, value interface{}, ttl time.Duration) error
}

// Get retrieves a value from the cache
func (m *MockCacheClient) Get(key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	if m.Data != nil {
		if v, ok := m.Data[key]; ok {
			return v, nil
		}
	}
	return "", nil
}

// Set stores a value in the cache with a TTL
func (m *MockCacheClient) Set(key string, value interface{}, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}
	if m.Data == nil {
		m.Data = make(map[string]string)
	}
	if s, ok := value.(string); ok {
		m.Data[key] = s
	}
	return nil
}

// Delete removes a key from the cache
func (m *MockCacheClient) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	if m.Data != nil {
		delete(m.Data, key)
	}
	return nil
}

// Exists checks if a key exists in the cache
func (m *MockCacheClient) Exists(key string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(key)
	}
	if m.Data != nil {
		_, ok := m.Data[key]
		return ok, nil
	}
	return false, nil
}

// Close closes the cache connection
func (m *MockCacheClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Ping checks if the cache connection is alive
func (m *MockCacheClient) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

// GetJSON retrieves a JSON value and unmarshals it into dest
func (m *MockCacheClient) GetJSON(key string, dest interface{}) error {
	if m.GetJSONFunc != nil {
		return m.GetJSONFunc(key, dest)
	}
	return nil
}

// SetJSON stores a value as JSON in the cache
func (m *MockCacheClient) SetJSON(key string, value interface{}, ttl time.Duration) error {
	if m.SetJSONFunc != nil {
		return m.SetJSONFunc(key, value, ttl)
	}
	return nil
}

// =============================================================================
// Mock Crawler Client
// =============================================================================

// MockCrawlerClient is a mock implementation of crawlers.CrawlerClient
type MockCrawlerClient struct {
	VisitedURLs       []string
	VisitFunc         func(url string) error
	VisitMultipleFunc func(urls []string) error
	WaitFunc          func()
	OnHTMLFunc        func(selector string, handler func(e *colly.HTMLElement))
	OnRequestFunc     func(handler func(r *colly.Request))
	OnResponseFunc    func(handler func(r *colly.Response))
	OnErrorFunc       func(handler func(r *colly.Response, err error))
	OnScrapedFunc     func(handler func(r *colly.Response))
	SetUserAgentFunc  func(ua string)
	SetHeadersFunc    func(headers map[string]string)
}

// Visit starts crawling from the given URL
func (m *MockCrawlerClient) Visit(url string) error {
	m.VisitedURLs = append(m.VisitedURLs, url)
	if m.VisitFunc != nil {
		return m.VisitFunc(url)
	}
	return nil
}

// VisitMultiple visits multiple URLs
func (m *MockCrawlerClient) VisitMultiple(urls []string) error {
	m.VisitedURLs = append(m.VisitedURLs, urls...)
	if m.VisitMultipleFunc != nil {
		return m.VisitMultipleFunc(urls)
	}
	return nil
}

// Wait waits for all async requests to complete
func (m *MockCrawlerClient) Wait() {
	if m.WaitFunc != nil {
		m.WaitFunc()
	}
}

// OnHTML registers a callback for HTML elements matching the selector
func (m *MockCrawlerClient) OnHTML(selector string, handler func(e *colly.HTMLElement)) {
	if m.OnHTMLFunc != nil {
		m.OnHTMLFunc(selector, handler)
	}
}

// OnRequest registers a callback before a request is made
func (m *MockCrawlerClient) OnRequest(handler func(r *colly.Request)) {
	if m.OnRequestFunc != nil {
		m.OnRequestFunc(handler)
	}
}

// OnResponse registers a callback after a response is received
func (m *MockCrawlerClient) OnResponse(handler func(r *colly.Response)) {
	if m.OnResponseFunc != nil {
		m.OnResponseFunc(handler)
	}
}

// OnError registers a callback when an error occurs
func (m *MockCrawlerClient) OnError(handler func(r *colly.Response, err error)) {
	if m.OnErrorFunc != nil {
		m.OnErrorFunc(handler)
	}
}

// OnScraped registers a callback after a page is scraped
func (m *MockCrawlerClient) OnScraped(handler func(r *colly.Response)) {
	if m.OnScrapedFunc != nil {
		m.OnScrapedFunc(handler)
	}
}

// SetUserAgent sets the user agent
func (m *MockCrawlerClient) SetUserAgent(ua string) {
	if m.SetUserAgentFunc != nil {
		m.SetUserAgentFunc(ua)
	}
}

// SetHeaders sets custom headers for requests
func (m *MockCrawlerClient) SetHeaders(headers map[string]string) {
	if m.SetHeadersFunc != nil {
		m.SetHeadersFunc(headers)
	}
}

// Ensure mocks implement the interfaces
var (
	_ interface {
		GetDB() *gorm.DB
		Create(value interface{}) error
		Find(dest interface{}, conds ...interface{}) error
		First(dest interface{}, conds ...interface{}) error
		Update(model interface{}, column string, value interface{}) error
		Updates(model interface{}, values interface{}) error
		Delete(value interface{}, conds ...interface{}) error
		Migrate(models ...interface{}) error
		Ping() error
		Close() error
		Transaction(fn func(*gorm.DB) error) error
	} = (*MockDatabaseClient)(nil)
)
