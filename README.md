# Golwarc - Go Web Crawler Master

[![CI](https://github.com/alonecandies/golwarc/actions/workflows/ci.yml/badge.svg)](https://github.com/alonecandies/golwarc/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go-based crawler master framework with multiple database clients, caching systems, message queues, and various crawler implementations.

## Architecture

![Golwarc Architecture](docs/architecture.png)

The architecture follows a clean dependency injection pattern:

- **main.go** - Entry point that initializes the DI container
- **inject/inject.go** - Dependency injection container that conditionally initializes all packages
- **services/** - Business logic layer that receives injected dependencies
- **Packages** - Independent modules for cache, database, crawlers, message queues, etc.

For detailed architecture documentation, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Documentation

Comprehensive documentation is available to help you understand and use Golwarc effectively:

- **[Architecture](ARCHITECTURE.md)** - Detailed architecture and dependency injection pattern explanation
- **[Performance](PERFORMANCE.md)** - Performance tips, optimization strategies, and benchmarks
- **[Security](SECURITY.md)** - Security best practices and guidelines
- **[Contributing](CONTRIBUTING.md)** - Guidelines for contributing to the project
- **[Changelog](CHANGELOG.md)** - Version history and release notes
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions

## Features

### 🗄️ Database Support

- **MySQL** - Full GORM integration with connection pooling
- **PostgreSQL** - GORM-based with transaction support
- **ClickHouse** - Optimized for analytical workloads
- **BigTable** - Google Cloud BigTable client

### 💾 Caching

- **LRU Cache** - In-memory LRU cache with configurable size
- **Redis** - Full-featured Redis client with TTL, JSON support, and atomic operations

### ⚙️ Configuration

- **Viper** - YAML/JSON configuration with environment variable overrides
- Structured configuration for all components
- Hot-reload support

### 📨 Message Queues

- **Kafka** - Producer and consumer with batch operations
- **RabbitMQ** - Full-featured client with exchange/queue management

### 🔄 Workflow Orchestration

- **Temporal** - Workflow and activity management
- Signal and query support
- Worker creation and management

### 🕷️ Crawler Implementations

#### Static Content Crawlers

- **Colly** - Fast and elegant scraper framework
- **Spider** - Custom crawler using goquery/cascadia
- **Soup** - Simple HTML parser

#### Dynamic Content Crawlers (JavaScript Support)

- **Selenium** - WebDriver-based browser automation
- **Playwright** - Multi-browser support (Chromium, Firefox, WebKit)
- **Puppeteer** - Chrome DevTools Protocol via chromedp
- **Ferret** - Declarative web scraping with FQL

### 📊 Models

Pre-built models for common scraping scenarios:

- **Page** - Web page data
- **Product** - E-commerce products
- **Article** - News articles and blog posts

## Installation

```bash
git clone https://github.com/alonecandies/golwarc
cd golwarc
go mod download
```

## Configuration

Copy the example configuration file and customize it:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Edit `configs/config.yaml` with your database credentials, cache settings, and other configurations.

## Quick Start

### 1. Initialize Logger

```go
package main

import "github.com/alonecandies/golwarc/logger"

func main() {
    // Initialize logger
    logger.InitDefaultLogger()
    defer logger.Sync()

    log := logger.GetLogger()
    log.Info("Hello from Golwarc!")
}
```

### 2. Load Configuration

```go
import "github.com/alonecandies/golwarc/configs"

config, err := configs.LoadConfig("configs/config.yaml")
if err != nil {
    panic(err)
}
```

### 3. Use Cache

```go
import "github.com/alonecandies/golwarc/cache"

// LRU Cache
lruCache, err := cache.NewLRUCache(1000)
lruCache.Set("key", "value")
value, exists := lruCache.Get("key")

// Redis Cache
redisClient, err := cache.NewRedisClient(cache.RedisConfig{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
redisClient.Set("key", "value", 10*time.Minute)
```

### 4. Database Operations

```go
import "github.com/alonecandies/golwarc/database"

// MySQL
mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: "password",
    Database: "golwarc",
})
defer mysqlClient.Close()

// Auto-migrate models
mysqlClient.Migrate(&models.Page{}, &models.Product{}, &models.Article{})
```

### 5. Web Crawling Examples

#### Using Colly (Static Content)

```go
import "github.com/alonecandies/golwarc/crawlers"

client := crawlers.NewDefaultCollyClient()

client.OnHTML("h1", func(e *colly.HTMLElement) {
    fmt.Println("Title:", e.Text)
})

client.OnHTML("a[href]", func(e *colly.HTMLElement) {
    link := e.Attr("href")
    fmt.Println("Link:", link)
    e.Request.Visit(link)
})

client.Visit("https://example.com")
client.Wait()
```

#### Using Playwright (Dynamic Content)

```go
import "github.com/alonecandies/golwarc/crawlers"

client, err := crawlers.NewPlaywrightClient(crawlers.PlaywrightConfig{
    BrowserType: "chromium",
    Headless:    true,
})
defer client.Close()

client.Navigate("https://example.com")
client.Click("button#load-more")
client.WaitForSelector(".dynamic-content")

content, err := client.GetContent()
fmt.Println(content)
```

#### Using Ferret FQL (Declarative)

```go
import "github.com/alonecandies/golwarc/crawlers"

client, err := crawlers.NewDefaultFerretClient()

// Extract all links
links, err := client.ExtractLinks("https://example.com")

// Extract products
products, err := client.ExampleExtractProducts("https://shop.example.com")
```

### 6. Message Queue Operations

#### Kafka

```go
import "github.com/alonecandies/golwarc/message-queue"

producer := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "crawl-results",
})
defer producer.Close()

producer.Produce(ctx, []byte("key"), []byte("message"))
```

#### RabbitMQ

```go
import "github.com/alonecandies/golwarc/message-queue"

client, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
    URL: "amqp://guest:guest@localhost:5672/",
})
defer client.Close()

client.DeclareQueue("tasks", true)
client.Publish(ctx, "tasks", []byte("task data"))
```

### 7. Temporal Workflow

```go
import "github.com/alonecandies/golwarc/libs"

temporalClient, err := libs.NewTemporalClient(libs.TemporalConfig{
    HostPort:  "localhost:7233",
    Namespace: "default",
})
defer temporalClient.Close()

// Start a workflow
run, err := temporalClient.StartWorkflow(ctx, "workflow-id", "task-queue", "workflowType", args)
```

## Usage Examples

### Complete Crawling Pipeline

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/alonecandies/golwarc"
    "github.com/alonecandies/golwarc/cache"
    "github.com/alonecandies/golwarc/configs"
    "github.com/alonecandies/golwarc/crawlers"
    "github.com/alonecandies/golwarc/database"
    "github.com/alonecandies/golwarc/models"
)

func main() {
    // Initialize logger
    golwarc.InitDefaultLogger()
    defer golwarc.Sync()

    // Load config
    config, _ := configs.LoadConfig("configs/config.yaml")

    // Setup cache
    redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
        Addr:     config.Cache.Redis.Addr,
        Password: config.Cache.Redis.Password,
        DB:       config.Cache.Redis.DB,
    })
    defer redisClient.Close()

    // Setup database
    mysqlClient, _ := database.NewMySQLClient(database.MySQLConfig{
        Host:     config.Database.MySQL.Host,
        Port:     config.Database.MySQL.Port,
        User:     config.Database.MySQL.User,
        Password: config.Database.MySQL.Password,
        Database: config.Database.MySQL.Database,
    })
    defer mysqlClient.Close()

    // Migrate models
    mysqlClient.Migrate(&models.Page{})

    // Create crawler
    crawler := crawlers.NewDefaultCollyClient()

    crawler.OnHTML("title", func(e *colly.HTMLElement) {
        // Check cache
        cached, _ := redisClient.Exists(e.Request.URL.String())
        if cached {
            return
        }

        // Save to database
        page := &models.Page{
            URL:    e.Request.URL.String(),
            Title:  e.Text,
            Status: 200,
        }
        mysqlClient.Create(page)

        // Cache the result
        redisClient.Set(e.Request.URL.String(), "processed", 24*time.Hour)

        fmt.Printf("Saved: %s\n", e.Text)
    })

    crawler.Visit("https://example.com")
    crawler.Wait()
}
```

## Testing

```bash
# Run all tests
go test ./tests/...

# Run with verbose output
go test -v ./tests/...

# Run with coverage
go test -cover ./tests/...

# Run specific test package
go test ./tests/cache/
go test ./tests/configs/
go test ./tests/models/
```

## Project Structure

```
golwarc/
├── cache/              # Cache implementations
│   ├── lru.go
│   └── redis.go
├── configs/            # Configuration management
│   ├── config.go
│   └── config.example.yaml
├── crawlers/           # Crawler implementations
│   ├── colly.go
│   ├── spider.go
│   ├── soup.go
│   ├── selenium.go
│   ├── playwright.go
│   ├── puppeteer.go
│   └── ferret.go
├── database/           # Database clients
│   ├── mysql.go
│   ├── postgresql.go
│   ├── clickhouse.go
│   └── bigtable.go
├── docker/             # Docker configuration
│   ├── Dockerfile
│   └── docker-compose.yaml
├── libs/               # Third-party integrations
│   └── temporal.go
├── logger/             # Logging configuration
│   └── logger.go
├── message-queue/      # Message queue clients
│   ├── kafka.go
│   └── rabbitmq.go
├── models/             # Data models
│   ├── page.go
│   ├── product.go
│   └── article.go
├── scripts/            # Development scripts
│   └── dev.sh
├── tests/              # Test suite
│   ├── cache/
│   │   └── cache_test.go
│   ├── configs/
│   │   └── config_test.go
│   └── models/
│       └── models_test.go
├── .env.example        # Environment variables template
├── config.example.yaml # Configuration template
├── go.mod
├── go.sum
├── main.go             # Demo application
└── README.md
```

## External Dependencies

### Browser Automation Tools

For Selenium, Playwright, and Puppeteer clients to work, you need:

#### Selenium

- Download ChromeDriver or GeckoDriver
- Or use a remote Selenium server (e.g., Selenium Grid)

#### Playwright

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright install
```

#### Puppeteer (chromedp)

- Requires Chrome/Chromium installed on the system

### Database Services

Ensure you have the following services running:

- MySQL (port 3306)
- PostgreSQL (port 5432)
- ClickHouse (port 9000)
- Redis (port 6379)
- Kafka (port 9092)
- RabbitMQ (port 5672)
- Temporal (port 7233)

Use Docker Compose for easy setup:

```yaml
version: "3.8"
services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: golwarc

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"

  # Add other services as needed
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance Tips

1. **Use connection pooling** - All database clients have built-in connection pooling
2. **Enable caching** - Use Redis to cache frequently accessed data
3. **Batch operations** - Use batch inserts for better throughput
4. **Rate limiting** - Configure delays in crawlers to avoid overwhelming targets
5. **Concurrent crawling** - Use Colly's async mode or Spider's concurrency settings

## FAQ

### General Questions

**Q: What Go version is required?**  
A: Go 1.25 or later is required.

**Q: Can I use this in production?**  
A: Yes! Golwarc is production-ready with comprehensive testing and error handling.

**Q: Which crawler should I use?**  
A:

- **Static sites**: Use Colly (fastest, most efficient)
- **JavaScript-heavy sites**: Use Playwright or Puppeteer
- **Complex scraping logic**: Use Ferret with FQL
- **Browser compatibility testing**: Use Selenium

**Q: Do I need all the external services (Redis, MySQL, etc.)?**  
A: No! Golwarc uses conditional initialization. Only configure the services you need.

### Configuration

**Q: How do I override config with environment variables?**  
A: Use the `GOLWARC_` prefix:

```bash
export GOLWARC_DATABASE_MYSQL_HOST=production.db
export GOLWARC_CACHE_REDIS_ADDR=redis.example.com:6379
```

**Q: Can I use environment variables exclusively?**  
A: Yes, either via `.env` file or export commands. The DI container will use defaults if config file is missing.

### Database

**Q: Can I use multiple databases simultaneously?**  
A: Yes! Configure all databases you need in `config.yaml` and they'll all be initialized.

**Q: How do I migrate to a new database schema?**  
A: Use GORM's AutoMigrate:

```go
container.MySQLClient.Migrate(&models.NewModel{})
```

**Q: What about database connection pooling?**  
A: All database clients have connection pooling pre-configured with sensible defaults (10 idle, 100 max open connections).

### Crawling

**Q: How do I avoid getting blocked while crawling?**  
A:

1. Add delays between requests
2. Rotate user agents
3. Use proxies (configure in crawler settings)
4. Respect robots.txt
5. Implement rate limiting

**Q: Can I crawl sites that require login?**  
A: Yes, use Playwright or Selenium to handle authentication flows.

**Q: How do I handle AJAX/dynamic content?**  
A: Use Playwright, Puppeteer, or Ferret instead of Colly/Soup.

### Caching

**Q: When should I use LRU vs Redis cache?**  
A:

- **LRU**: Single instance, development, simple caching needs
- **Redis**: Distributed systems, production, shared cache across services

**Q: Can I use both caches together?**  
A: Yes! Use LRU as L1 cache and Redis as L2 for multi-tier caching.

### Testing

**Q: Why do tests skip?**  
A: Tests skip when external services aren't available. Use `make docker-up` to start all services.

**Q: How do I run only unit tests?**  
A: Unit tests that don't require external services will run:

```bash
go test ./tests/configs/
go test ./tests/models/
```

### Performance

**Q: How many concurrent crawlers can I run?**  
A: Depends on your resources. Start with 5-10 and monitor CPU/memory. Use message queues (Kafka/RabbitMQ) for distributed crawling.

**Q: How do I improve crawling speed?**  
A:

1. Use Colly's async mode
2. Increase concurrency settings
3. Enable caching to avoid re-crawls
4. Use connection pooling
5. Batch database inserts

## Troubleshooting

For common issues and solutions, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

Quick links:

- [Database connection issues](TROUBLESHOOTING.md#database-connection-issues)
- [Crawler issues](TROUBLESHOOTING.md#crawler-issues)
- [Cache issues](TROUBLESHOOTING.md#cache-issues)
- [Test issues](TROUBLESHOOTING.md#test-issues)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [Colly](https://github.com/gocolly/colly) - Fast web scraping framework
- [goquery](https://github.com/PuerkitoBio/goquery) - jQuery-like HTML parsing
- [Playwright](https://github.com/playwright-community/playwright-go) - Browser automation
- [chromedp](https://github.com/chromedp/chromedp) - Chrome DevTools Protocol
- [Ferret](https://github.com/MontFerret/ferret) - Declarative web scraping
- [GORM](https://gorm.io/) - ORM library for Go
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Fast structured logging

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/alonecandies/golwarc).
