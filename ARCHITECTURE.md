# Golwarc Architecture - Dependency Injection Pattern

## Overview

Golwarc uses a clean dependency injection pattern to manage all application dependencies and services.

## Architecture Layers

```
┌─────────────────────────────────────────┐
│          main.go (Entry Point)          │
│  - Loads config                         │
│  - Creates DI container                 │
│  - Runs services                        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      inject/inject.go (DI Container)    │
│  - Reads configuration                  │
│  - Initializes all packages             │
│  - Manages lifecycle                    │
└──────────────┬──────────────────────────┘
               │
               ├──────────────┬──────────────┬─────────────┐
               ▼              ▼              ▼             ▼
         ┌─────────┐    ┌─────────┐   ┌──────────┐  ┌──────────┐
         │  cache  │    │database │   │  logger  │  │   etc    │
         └─────────┘    └─────────┘   └──────────┘  └──────────┘
               │              │              │             │
               └──────────────┴──────────────┴─────────────┘
                              │
                              ▼
                   ┌──────────────────────┐
                   │  services/           │
                   │  - CrawlerService    │
                   │  - (Other services)  │
                   └──────────────────────┘
```

## Components

### 1. Entry Point: `main.go`

Responsibilities:

- Initialize DI container
- Display available services
- Run service demos
- Handle graceful shutdown

```go
container, err := inject.NewContainer("config.yaml")
defer container.Close()

crawlerService := services.NewCrawlerService(
    container.Logger,
    container.RedisClient,
    container.MySQLClient,
)
```

### 2. Dependency Injection: `inject/inject.go`

Responsibilities:

- Read configuration file
- Initialize packages conditionally
- Manage dependencies lifecycle
- Provide centralized access to all services

**Container Structure:**

```go
type Container struct {
    Logger       *zap.Logger
    Config       *configs.Config
    LRUCache     *cache.LRUCache
    RedisClient  *cache.RedisClient
    MySQLClient  *database.MySQLClient
    PGClient     *database.PostgresClient
    CHClient     *database.ClickHouseClient
    KafkaClient  *messagequeue.KafkaProducer
    RabbitClient *messagequeue.RabbitMQClient
}
```

**Conditional Initialization:**

- Only initializes services that are configured
- Logs warnings for services that fail to initialize
- Continues execution even if optional services are unavailable

### 3. Service Layer: `services/`

Responsibilities:

- Business logic implementation
- Uses injected dependencies
- Independent and testable
- Stateless where possible

**Example: CrawlerService**

```go
type CrawlerService struct {
    logger     *zap.Logger
    redisCache *cache.RedisClient
    mysqlDB    *database.MySQLClient
    crawler    *crawlers.CollyClient
}
```

Methods:

- `Initialize()` - Sets up database schema
- `CrawlAndStore(url)` - Crawls, caches, and persists
- `GetStats()` - Returns statistics
- `GetRecentPages(limit)` - Retrieves recent data

### 4. Package Layer

Independent packages that can be injected:

- **cache/** - LRU and Redis caching
- **database/** - MySQL, PostgreSQL, ClickHouse, BigTable
- **logger/** - Structured logging with Zap
- **configs/** - Configuration management
- **message-queue/** - Kafka and RabbitMQ
- **crawlers/** - Various crawler implementations
- **models/** - Data models

## Configuration-Driven Initialization

The DI container reads `config.yaml` and only initializes configured services:

```yaml
cache:
  lru:
    size: 1000 # LRU will be initialized
  redis:
    addr: "localhost:6379" # Redis will be initialized

database:
  mysql:
    host: "localhost" # MySQL will be initialized
  postgresql:
    host: "" # PostgreSQL will NOT be initialized (empty)
```

## Workflow

1. **Startup**

   ```
   dev.sh → Export .env → Run main.go
   ```

2. **Initialization**

   ```
   main.go → inject.NewContainer() → Initialize configured packages
   ```

3. **Service Creation**

   ```
   Container → Inject dependencies → Create service instances
   ```

4. **Execution**

   ```
   Service methods → Use injected packages → Perform operations
   ```

5. **Shutdown**
   ```
   container.Close() → Close all connections → Cleanup resources
   ```

## Benefits

✅ **Testability**: Easy to mock dependencies  
✅ **Flexibility**: Services configured via YAML  
✅ **Maintainability**: Clear separation of concerns  
✅ **Scalability**: Easy to add new services  
✅ **Resilience**: Graceful degradation if services unavailable

## Adding a New Service

1. **Create service file** in `services/`:

   ```go
   type MyService struct {
       logger *zap.Logger
       db     *database.MySQLClient
   }

   func NewMyService(logger *zap.Logger, db *database.MySQLClient) *MyService {
       return &MyService{logger: logger, db: db}
   }
   ```

2. **Use in main.go**:

   ```go
   myService := services.NewMyService(
       container.Logger,
       container.MySQLClient,
   )
   ```

3. **Add methods to service**:
   ```go
   func (s *MyService) DoSomething() error {
       s.logger.Info("Doing something...")
       // Use s.db to interact with database
       return nil
   }
   ```

## Running the Application

```bash
# Simple run
./scripts/dev.sh

# Or with custom config
CONFIG_PATH=custom.yaml go run main.go

# Or with environment override
MYSQL_HOST=production.db go run main.go
```

## Environment Variables

The `.env` file can override configuration:

```bash
MYSQL_HOST=localhost
MYSQL_PORT=3306
REDIS_ADDR=localhost:6379
LOG_LEVEL=debug
```

These are loaded by `dev.sh` before running main.go.

## Testing Services

Services can be easily tested by mocking dependencies:

```go
func TestCrawlerService(t *testing.T) {
    mockLogger := zaptest.NewLogger(t)
    mockCache := &MockRedisClient{}
    mockDB := &MockMySQLClient{}

    service := NewCrawlerService(mockLogger, mockCache, mockDB)

    err := service.CrawlAndStore("https://example.com")
    assert.NoError(t, err)
}
```

---

This architecture provides a clean, maintainable, and scalable foundation for the crawler master application.
