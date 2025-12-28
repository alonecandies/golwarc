# Troubleshooting Guide

This guide helps you resolve common issues when using Golwarc.

## Installation Issues

### Issue: `go mod download` fails

**Symptom:** Dependency download errors or timeout

**Solution:**

```bash
# Clear module cache
go clean -modcache

# Set Go proxy
export GOPROXY=https://proxy.golang.org,direct

# Retry download
go mod download
```

### Issue: Go version mismatch

**Symptom:** `go: module requires go 1.25 (running go 1.xx)`

**Solution:**

```bash
# Check your Go version
go version

# Update to Go 1.25 or later
# Visit: https://golang.org/dl/
```

---

## Database Connection Issues

### Issue: MySQL connection refused

**Symptom:** `dial tcp 127.0.0.1:3306: connect: connection refused`

**Solution:**

```bash
# 1. Check if MySQL is running
mysql --version
mysqladmin ping

# 2. Start MySQL (macOS)
brew services start mysql

# 3. Start MySQL (Linux)
sudo systemctl start mysql

# 4. Verify credentials in config.yaml
```

### Issue: PostgreSQL authentication failed

**Symptom:** `FATAL: password authentication failed`

**Solution:**

```bash
# 1. Reset PostgreSQL password
sudo -u postgres psql
ALTER USER postgres PASSWORD 'your_password';

# 2. Update config.yaml with correct credentials

# 3. Check pg_hba.conf authentication method
# Should have: host all all 127.0.0.1/32 md5
```

### Issue: ClickHouse timeout error

**Symptom:** `read timeout: time: missing unit in duration`

**Solution:**
Update `config.yaml`:

```yaml
database:
  clickhouse:
    read_timeout: "10s" # Add 's' suffix for seconds
    write_timeout: "10s"
```

---

## Cache Issues

### Issue: Redis connection refused

**Symptom:** `dial tcp 127.0.0.1:6379: connect: connection refused`

**Solution:**

```bash
# 1. Check if Redis is running
redis-cli ping

# 2. Start Redis (macOS)
brew services start redis

# 3. Start Redis (Linux)
sudo systemctl start redis

# 4. Start Redis (Docker)
docker run -d -p 6379:6379 redis:7-alpine
```

### Issue: LRU cache full

**Symptom:** Cache evicting entries too frequently

**Solution:**
Increase cache size in `config.yaml`:

```yaml
cache:
  lru:
    size: 10000 # Increase from 1000
```

---

## Crawler Issues

### Issue: Playwright browser not found

**Symptom:** `browser executable not found`

**Solution:**

```bash
# Install Playwright browsers
go run github.com/playwright-community/playwright-go/cmd/playwright install

# Or install specific browser
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

### Issue: Selenium WebDriver error

**Symptom:** `WebDriver not found` or `ChromeDriver executable not found`

**Solution:**

```bash
# Download ChromeDriver
# Visit: https://chromedriver.chromium.org/downloads

# Or use WebDriver manager
go get github.com/tebeka/selenium
```

### Issue: Rate limiting / IP blocked

**Symptom:** 429 Too Many Requests or connection refused

**Solution:**
Add delays in crawler configuration:

```go
crawler := crawlers.NewCollyClient(crawlers.CollyConfig{
    Delay: 2 * time.Second,  // Add 2 second delay
    RandomDelay: 1 * time.Second,
})
```

### Issue: JavaScript not rendering

**Symptom:** Using Colly/Soup but page content is empty

**Solution:**
Switch to a dynamic crawler:

```go
// Instead of Colly, use Playwright for JS-heavy sites
client, err := crawlers.NewPlaywrightClient(crawlers.PlaywrightConfig{
    BrowserType: "chromium",
    Headless:    true,
})
```

---

## Message Queue Issues

### Issue: Kafka connection failed

**Symptom:** `Failed to connect to broker`

**Solution:**

```bash
# 1. Start Kafka (Docker)
docker run -d -p 9092:9092 confluentinc/cp-kafka:latest

# 2. Verify broker is listening
nc -zv localhost 9092

# 3. Check config.yaml brokers list
```

### Issue: RabbitMQ authentication error

**Symptom:** `ACCESS_REFUSED - Login was refused`

**Solution:**
Update `config.yaml`:

```yaml
message_queue:
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/" # Check credentials
```

---

## Build Issues

### Issue: Build fails with missing packages

**Symptom:** `package xxx is not in GOROOT`

**Solution:**

```bash
# Update dependencies
go mod tidy

# Download missing packages
go mod download

# Verify go.mod
go mod verify
```

### Issue: Linter errors

**Symptom:** `golangci-lint` reports errors

**Solution:**

```bash
# Run linter to see errors
make lint

# Auto-fix formatting issues
make fmt

# Update golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Test Issues

### Issue: Tests skip with "service not available"

**Symptom:** Tests marked as SKIP

**Solution:**
This is expected behavior. Tests skip when external services (Redis, MySQL) aren't available.

To run full tests:

```bash
# Start all services with Docker Compose
make docker-up

# Run tests
make test

# Stop services
make docker-down
```

### Issue: Race detector warnings

**Symptom:** `WARNING: DATA RACE`

**Solution:**

```bash
# Run tests with race detector to identify
go test -race ./tests/...

# Fix race conditions using:
# - sync.Mutex for shared state
# - sync.RWMutex for read-heavy operations
# - atomic operations for counters
```

---

## Configuration Issues

### Issue: Config file not found

**Symptom:** `Failed to load config: open config.yaml: no such file or directory`

**Solution:**

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit with your settings
nano config.yaml
```

### Issue: Environment variables not working

**Symptom:** Config not reading from env vars

**Solution:**
Viper automatically reads env vars with prefix. Use:

```bash
export GOLWARC_DATABASE_MYSQL_HOST=localhost
export GOLWARC_DATABASE_MYSQL_PORT=3306
```

---

## Performance Issues

### Issue: Slow database queries

**Solution:**

1. Add database indexes:

```go
db.Model(&models.Page{}).AddIndex("idx_url", "url")
db.Model(&models.Page{}).AddIndex("idx_created_at", "created_at")
```

2. Use batch operations:

```go
// Instead of looping Create()
db.CreateInBatches(pages, 100)
```

3. Increase connection pool:

```yaml
database:
  mysql:
    max_open_conns: 200
    max_idle_conns: 20
```

### Issue: High memory usage

**Solution:**

1. Reduce LRU cache size
2. Limit crawler concurrency
3. Use streaming for large datasets:

```go
rows, _ := db.Raw("SELECT * FROM pages").Rows()
defer rows.Close()

for rows.Next() {
    // Process one at a time
}
```

---

## Docker Issues

### Issue: Docker Compose fails to start

**Symptom:** `Cannot start service xxx`

**Solution:**

```bash
# Check Docker is running
docker info

# View service logs
docker-compose -f docker/docker-compose.yaml logs

# Restart services
docker-compose -f docker/docker-compose.yaml restart
```

---

## Getting Help

If you encounter issues not covered here:

1. **Check logs** - Enable debug logging:

   ```yaml
   log:
     level: debug
   ```

2. **Search Issues** - Check [GitHub Issues](https://github.com/alonecandies/golwarc/issues)

3. **Create Issue** - Include:

   - Go version (`go version`)
   - OS (`uname -a`)
   - Error message
   - Config file (redact credentials)
   - Steps to reproduce

4. **Community** - Ask in Discussions or submit a PR with fixes!

---

**Last Updated:** December 28, 2025
