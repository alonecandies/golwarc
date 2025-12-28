package main

import (
	"fmt"
	stdlog "log"

	"github.com/alonecandies/golwarc/inject"
	"github.com/alonecandies/golwarc/services"
	"go.uber.org/zap"
)

func main() {
	// Initialize dependency injection container
	container, err := inject.NewContainer("config.yaml")
	if err != nil {
		stdlog.Fatalf("Failed to initialize container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			stdlog.Printf("Warning: error closing container: %v", err)
		}
	}()

	log := container.Logger
	log.Info("==============================================")
	log.Info("Golwarc Crawler Master - Dependency Injection Demo")
	log.Info("==============================================")

	// Check which services are available
	log.Info("Available services:")
	if container.LRUCache != nil {
		log.Info("  ✓ LRU Cache")
	}
	if container.RedisClient != nil {
		log.Info("  ✓ Redis Cache")
	}
	if container.MySQLClient != nil {
		log.Info("  ✓ MySQL Database")
	}
	if container.PGClient != nil {
		log.Info("  ✓ PostgreSQL Database")
	}
	if container.CHClient != nil {
		log.Info("  ✓ ClickHouse Database")
	}
	if container.KafkaClient != nil {
		log.Info("  ✓ Kafka Producer")
	}
	if container.RabbitClient != nil {
		log.Info("  ✓ RabbitMQ Client")
	}

	// Demonstrate service usage
	if container.RedisClient != nil && container.MySQLClient != nil {
		runCrawlerDemo(container)
	} else {
		log.Warn("Crawler demo requires Redis and MySQL to be configured")
		log.Info("Please configure database.mysql and cache.redis in config.yaml")
	}

	log.Info("==============================================")
	log.Info("Demo completed successfully!")
	log.Info("==============================================")
}

func runCrawlerDemo(container *inject.Container) {
	log := container.Logger
	log.Info("")
	log.Info("--- Crawler Service Demo ---")
	log.Info("")

	// Create crawler service with injected dependencies
	crawlerService := services.NewCrawlerService(
		container.Logger,
		container.RedisClient,
		container.MySQLClient,
	)

	// Initialize service (migrate database)
	log.Info("Initializing crawler service...")
	if err := crawlerService.Initialize(); err != nil {
		log.Fatal("Failed to initialize crawler service", zap.Error(err))
	}

	// Crawl example URLs
	urls := []string{
		"https://example.com",
		"https://example.org",
	}

	for _, url := range urls {
		log.Info("Crawling URL...", zap.String("url", url))
		if err := crawlerService.CrawlAndStore(url); err != nil {
			log.Error("Failed to crawl URL", zap.String("url", url), zap.Error(err))
		}
	}

	// Get statistics
	log.Info("")
	log.Info("Fetching crawler statistics...")
	stats, err := crawlerService.GetStats()
	if err != nil {
		log.Error("Failed to get stats", zap.Error(err))
	} else {
		log.Info("Crawler statistics:", zap.Any("stats", stats))
	}

	// Get recent pages
	log.Info("")
	log.Info("Fetching recent pages...")
	pages, err := crawlerService.GetRecentPages(5)
	if err != nil {
		log.Error("Failed to get recent pages", zap.Error(err))
	} else {
		log.Info(fmt.Sprintf("Retrieved %d recent pages:", len(pages)))
		for i := range pages {
			log.Info(fmt.Sprintf("  %d. %s - %s", i+1, pages[i].Title, pages[i].URL))
		}
	}

	log.Info("")
	log.Info("--- Crawler Service Demo Complete ---")
}
