# Performance Guidelines

This document outlines performance best practices, benchmarks, and Service Level Objectives (SLOs) for Golwarc.

## Service Level Objectives (SLOs)

### Crawler Performance

| Metric                | Target            | Measurement             |
| --------------------- | ----------------- | ----------------------- |
| Throughput            | ≥ 50 pages/second | Per crawler instance    |
| Request Latency (p95) | ≤ 2 seconds       | Time to fetch and parse |
| Request Latency (p99) | ≤ 5 seconds       | Time to fetch and parse |
| Success Rate          | ≥ 95%             | Non-error responses     |

### Cache Performance

| Metric                  | Target  | Measurement                 |
| ----------------------- | ------- | --------------------------- |
| Hit Ratio               | ≥ 80%   | Cache hits / total requests |
| Operation Latency (p95) | ≤ 10ms  | Redis operations            |
| Operation Latency (p99) | ≤ 50ms  | Redis operations            |
| Availability            | ≥ 99.9% | Cache uptime                |

### Database Performance

| Metric                      | Target  | Measurement              |
| --------------------------- | ------- | ------------------------ |
| Query Latency (p95)         | ≤ 100ms | SELECT queries           |
| Query Latency (p99)         | ≤ 500ms | SELECT queries           |
| Insert Latency (p95)        | ≤ 50ms  | Single row inserts       |
| Connection Pool Utilization | ≤ 80%   | Active connections / max |
| Transaction Success Rate    | ≥ 99.5% | Committed / total        |

### System Resources

| Metric           | Target   | Measurement            |
| ---------------- | -------- | ---------------------- |
| CPU Utilization  | ≤ 70%    | Average over 5 minutes |
| Memory Usage     | ≤ 80%    | RSS memory             |
| Goroutines       | ≤ 10,000 | Active goroutines      |
| File Descriptors | ≤ 1,000  | Open file descriptors  |

## Performance Best Practices

### 1. Connection Pooling

All database clients use connection pooling by default:

```go
// MySQL/PostgreSQL default settings
MaxIdleConns: 10
MaxOpenConns: 100
ConnMaxLifetime: 1 hour
```

**Recommendation**: Adjust based on your workload. For high-traffic scenarios, increase `MaxOpenConns` to 200-500.

### 2. Caching Strategy

Implement a two-tier caching approach:

1. **L1 Cache (LRU)**: In-memory, ultra-fast (< 1μs)
2. **L2 Cache (Redis)**: Distributed, fast (< 10ms)

```go
// Check L1 first
if val, found := lruCache.Get(key); found {
    return val
}

// Check L2
if val, err := redisClient.Get(key); err == nil {
    lruCache.Set(key, val) // Populate L1
    return val
}

// Cache miss - fetch from source
```

### 3. Rate Limiting

Configure rate limiting to prevent resource exhaustion:

```yaml
crawler:
  rate_limit:
    enabled: true
    requests_per_sec: 10 # Adjust based on target server tolerance
    max_concurrent: 5 # Concurrent requests
```

```go
limiter := libs.NewRateLimiter(libs.RateLimiterConfig{
    RequestsPerSecond: 10,
    Burst: 20,
})

// Before each request
if err := limiter.Wait(ctx); err != nil {
    return err
}
```

### 4. Batch Operations

Use batch operations for database inserts:

```go
// Bad: Individual inserts
for _, page := range pages {
    db.Create(page) // N queries
}

// Good: Batch insert
db.CreateInBatches(pages, 100) // N/100 queries
```

### 5. Concurrent Crawling

Use Colly's async mode for parallel crawling:

```go
crawler := colly.NewCollector(
    colly.Async(true),
)
crawler.Limit(&colly.LimitRule{
    DomainGlob:  "*",
    Parallelism: 5,
    Delay:       2 * time.Second,
})
```

### 6. Memory Management

- Set appropriate LRU cache size (don't cache everything)
- Use streaming for large responses
- Implement pagination for bulk queries

```go
// Bad: Load all results
var pages []Page
db.Find(&pages) // Could be millions

// Good: Paginate
db.Limit(100).Offset(0).Find(&pages)
```

## Running Benchmarks

Golwarc includes comprehensive benchmarks in `tests/benchmarks/`:

```bash
# Run all benchmarks
go test -bench=. ./tests/benchmarks/...

# Run with memory profiling
go test -bench=. -benchmem ./tests/benchmarks/...

# Run specific benchmark
go test -bench=BenchmarkRedisOperations ./tests/benchmarks/

# Save results for comparison
go test -bench=. ./tests/benchmarks/ | tee benchmark-results.txt
```

### Interpreting Results

```
BenchmarkRedisSet-8    100000    12000 ns/op    256 B/op    4 allocs/op
```

- `100000`: Number of iterations
- `12000 ns/op`: 12 microseconds per operation
- `256 B/op`: 256 bytes allocated per operation
- `4 allocs/op`: 4 memory allocations per operation

**Targets:**

- Redis operations: < 10,000 ns/op (10μs)
- LRU cache: < 1,000 ns/op (1μs)
- Database queries: < 100,000 ns/op (100μs)

## Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof -bench=. ./tests/benchmarks/
go tool pprof cpu.prof
```

In pprof:

```
top10        # Show top 10 functions by CPU
list funcName # Show source code
web          # Open in browser (requires graphviz)
```

### Memory Profiling

```bash
go test -memprofile=mem.prof -bench=. ./tests/benchmarks/
go tool pprof mem.prof
```

### Live Profiling

Add pprof to your application:

```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

Access profiles:

- CPU: `http://localhost:6060/debug/pprof/profile?seconds=30`
- Heap: `http://localhost:6060/debug/pprof/heap`
- Goroutines: `http://localhost:6060/debug/pprof/goroutine`

## Load Testing

### Using Apache Bench

```bash
# Test 1000 requests with 10 concurrent
ab -n 1000 -c 10 http://localhost:8080/crawl?url=example.com
```

### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 10000 -c 100 -q 10 http://localhost:8080/api/endpoint
```

### Expected Results

For a typical setup (4 CPU cores, 8GB RAM):

| Concurrent Users | Throughput | Avg Latency | p95 Latency |
| ---------------- | ---------- | ----------- | ----------- |
| 10               | 50 req/s   | 200ms       | 300ms       |
| 50               | 200 req/s  | 250ms       | 500ms       |
| 100              | 350 req/s  | 285ms       | 800ms       |
| 500              | 400 req/s  | 1.2s        | 3s          |

## Optimization Checklist

- [ ] Enable connection pooling for all databases
- [ ] Implement two-tier caching (LRU + Redis)
- [ ] Configure rate limiting for crawlers
- [ ] Use batch operations for bulk inserts
- [ ] Enable async mode for concurrent crawling
- [ ] Set appropriate cache TTLs
- [ ] Monitor Prometheus metrics
- [ ] Set up alerts for SLO violations
- [ ] Profile application under load
- [ ] Optimize database queries (add indices)

## Monitoring with Prometheus

Golwarc exports Prometheus metrics on `/metrics`:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: "golwarc"
    scrape_interval: 15s
    static_configs:
      - targets: ["localhost:8080"]
```

### Key Metrics

```promql
# Crawler throughput
rate(golwarc_crawler_requests_total[5m])

# Cache hit ratio
sum(rate(golwarc_cache_operations_total{operation="hit"}[5m])) /
sum(rate(golwarc_cache_operations_total[5m]))

# Database latency p95
histogram_quantile(0.95, golwarc_database_query_duration_seconds_bucket)

# Error rate
rate(golwarc_crawler_errors_total[5m])
```

## Alerting Rules

```yaml
groups:
  - name: golwarc_alerts
    rules:
      - alert: HighCrawlerErrorRate
        expr: rate(golwarc_crawler_errors_total[5m]) > 0.1
        annotations:
          summary: "High crawler error rate"

      - alert: LowCacheHitRatio
        expr: sum(rate(golwarc_cache_operations_total{status="hit"}[5m])) / sum(rate(golwarc_cache_operations_total[5m])) < 0.7
        annotations:
          summary: "Cache hit ratio below 70%"

      - alert: HighDatabaseLatency
        expr: histogram_quantile(0.95, golwarc_database_query_duration_seconds_bucket) > 0.5
        annotations:
          summary: "Database p95 latency exceeds 500ms"
```

## Troubleshooting Performance Issues

### High Latency

1. Check database query performance: `EXPLAIN ANALYZE SELECT ...`
2. Monitor connection pool utilization
3. Review cache hit ratios
4. Profile CPU usage with pprof

### High Memory Usage

1. Check goroutine leaks: `http://localhost:6060/debug/pprof/goroutine`
2. Review LRU cache size configuration
3. Look for memory leaks in crawlers (unclosed connections)
4. Analyze heap profile: `go tool pprof heap.prof`

### Low Throughput

1. Increase crawler concurrency
2. Reduce rate limit delays
3. Use async crawling mode
4. Scale horizontally (multiple instances)
5. Optimize database indices

---

**Last Updated:** December 28, 2025
