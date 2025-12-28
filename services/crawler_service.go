package services

import (
	"fmt"
	"time"

	"github.com/alonecandies/golwarc/cache"
	"github.com/alonecandies/golwarc/crawlers"
	"github.com/alonecandies/golwarc/database"
	"github.com/alonecandies/golwarc/models"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

// CrawlerService handles web crawling with caching and persistence
type CrawlerService struct {
	logger  *zap.Logger
	cache   cache.JSONCacheClient
	db      database.DatabaseClient
	crawler crawlers.CrawlerClient
}

// NewCrawlerService creates a new crawler service with injected dependencies
func NewCrawlerService(
	logger *zap.Logger,
	cacheClient cache.JSONCacheClient,
	dbClient database.DatabaseClient,
) *CrawlerService {
	return &CrawlerService{
		logger:  logger,
		cache:   cacheClient,
		db:      dbClient,
		crawler: crawlers.NewDefaultCollyClient(),
	}
}

// Initialize sets up the database schema
func (s *CrawlerService) Initialize() error {
	s.logger.Info("Initializing crawler service database schema")

	// Auto-migrate models
	if err := s.db.Migrate(&models.Page{}, &models.Product{}, &models.Article{}); err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
	}

	s.logger.Info("Database schema initialized successfully")
	return nil
}

// CrawlAndStore crawls a URL, caches the result, and stores in database
func (s *CrawlerService) CrawlAndStore(url string) error {
	s.logger.Info("Starting crawl", zap.String("url", url))

	// Check cache first
	cacheKey := fmt.Sprintf("page:%s", url)
	if s.cache != nil {
		cached, err := s.cache.Exists(cacheKey)
		if err == nil && cached {
			s.logger.Info("Page found in cache, skipping crawl", zap.String("url", url))
			return nil
		}
	}

	var crawledPage *models.Page
	var crawlErr error

	// Set up crawler callbacks
	s.crawler.OnHTML("html", func(e *colly.HTMLElement) {
		title := e.ChildText("title")
		if title == "" {
			title = "No title"
		}

		s.logger.Info("Page scraped",
			zap.String("url", url),
			zap.String("title", title))

		// Create page model
		crawledPage = &models.Page{
			URL:    url,
			Title:  title,
			Domain: e.Request.URL.Host,
			Status: 200,
			HTML:   string(e.Response.Body),
		}
	})

	s.crawler.OnError(func(r *colly.Response, err error) {
		crawlErr = err
		s.logger.Error("Crawl failed",
			zap.String("url", url),
			zap.Error(err))
	})

	// Visit the URL
	if err := s.crawler.Visit(url); err != nil {
		return fmt.Errorf("failed to visit URL: %w", err)
	}

	s.crawler.Wait()

	if crawlErr != nil {
		return crawlErr
	}

	if crawledPage == nil {
		return fmt.Errorf("no data extracted from URL")
	}

	// Save to database
	if err := s.db.Create(crawledPage); err != nil {
		s.logger.Error("Failed to save page to database",
			zap.String("url", url),
			zap.Error(err))
		return fmt.Errorf("failed to save to database: %w", err)
	}

	s.logger.Info("Page saved to database",
		zap.String("url", url),
		zap.Uint("page_id", crawledPage.ID))

	// Cache the result
	if s.cache != nil {
		if err := s.cache.SetJSON(cacheKey, crawledPage, 24*time.Hour); err != nil {
			s.logger.Warn("Failed to cache page",
				zap.String("url", url),
				zap.Error(err))
		} else {
			s.logger.Info("Page cached",
				zap.String("url", url),
				zap.Duration("ttl", 24*time.Hour))
		}
	}

	return nil
}

// GetStats returns crawler statistics
func (s *CrawlerService) GetStats() (map[string]interface{}, error) {
	s.logger.Info("Fetching crawler statistics")

	var pageCount int64
	if err := s.db.GetDB().Model(&models.Page{}).Count(&pageCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count pages: %w", err)
	}

	stats := map[string]interface{}{
		"total_pages_crawled": pageCount,
		"cache_enabled":       s.cache != nil,
		"database_connected":  s.db != nil,
	}

	s.logger.Info("Statistics retrieved", zap.Any("stats", stats))
	return stats, nil
}

// GetRecentPages retrieves the most recently crawled pages
func (s *CrawlerService) GetRecentPages(limit int) ([]models.Page, error) {
	var pages []models.Page

	err := s.db.GetDB().
		Order("created_at DESC").
		Limit(limit).
		Find(&pages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent pages: %w", err)
	}

	s.logger.Info("Retrieved recent pages", zap.Int("count", len(pages)))
	return pages, nil
}
