package libs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// Crawler metrics
	CrawlerRequestsTotal *prometheus.CounterVec
	CrawlerDuration      *prometheus.HistogramVec
	CrawlerErrorsTotal   *prometheus.CounterVec

	// Cache metrics
	CacheOperationsTotal *prometheus.CounterVec
	CacheDuration        *prometheus.HistogramVec

	// Database metrics
	DatabaseQueriesTotal  *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseErrorsTotal   *prometheus.CounterVec

	// System metrics
	ActiveConnections prometheus.Gauge
	HealthStatus      *prometheus.GaugeVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		// Crawler metrics
		CrawlerRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "golwarc_crawler_requests_total",
				Help: "Total number of crawler requests",
			},
			[]string{"crawler_type", "status"},
		),
		CrawlerDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "golwarc_crawler_duration_seconds",
				Help:    "Duration of crawler requests in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"crawler_type"},
		),
		CrawlerErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "golwarc_crawler_errors_total",
				Help: "Total number of crawler errors",
			},
			[]string{"crawler_type", "error_type"},
		),

		// Cache metrics
		CacheOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "golwarc_cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"cache_type", "operation", "status"},
		),
		CacheDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "golwarc_cache_duration_seconds",
				Help:    "Duration of cache operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"cache_type", "operation"},
		),

		// Database metrics
		DatabaseQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "golwarc_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"database_type", "operation"},
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "golwarc_database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
			},
			[]string{"database_type", "operation"},
		),
		DatabaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "golwarc_database_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"database_type", "error_type"},
		),

		// System metrics
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "golwarc_active_connections",
				Help: "Number of active connections",
			},
		),
		HealthStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "golwarc_health_status",
				Help: "Health status of services (1 = healthy, 0 = unhealthy)",
			},
			[]string{"service"},
		),
	}

	return metrics
}

// MetricsServer holds the HTTP server for metrics
type MetricsServer struct {
	server  *http.Server
	Metrics *Metrics
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	metrics := NewMetrics()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			// Log error but don't fail on write error for health check
			return
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &MetricsServer{
		server:  server,
		Metrics: metrics,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start() error {
	return ms.server.ListenAndServe()
}

// Stop gracefully stops the metrics server
func (ms *MetricsServer) Stop() error {
	return ms.server.Close()
}

// RecordCrawlerRequest records a crawler request
func (m *Metrics) RecordCrawlerRequest(crawlerType, status string) {
	m.CrawlerRequestsTotal.WithLabelValues(crawlerType, status).Inc()
}

// RecordCrawlerDuration records crawler duration
func (m *Metrics) RecordCrawlerDuration(crawlerType string, duration time.Duration) {
	m.CrawlerDuration.WithLabelValues(crawlerType).Observe(duration.Seconds())
}

// RecordCrawlerError records a crawler error
func (m *Metrics) RecordCrawlerError(crawlerType, errorType string) {
	m.CrawlerErrorsTotal.WithLabelValues(crawlerType, errorType).Inc()
}

// RecordCacheOperation records a cache operation
func (m *Metrics) RecordCacheOperation(cacheType, operation, status string) {
	m.CacheOperationsTotal.WithLabelValues(cacheType, operation, status).Inc()
}

// RecordCacheDuration records cache operation duration
func (m *Metrics) RecordCacheDuration(cacheType, operation string, duration time.Duration) {
	m.CacheDuration.WithLabelValues(cacheType, operation).Observe(duration.Seconds())
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(databaseType, operation string) {
	m.DatabaseQueriesTotal.WithLabelValues(databaseType, operation).Inc()
}

// RecordDatabaseDuration records database query duration
func (m *Metrics) RecordDatabaseDuration(databaseType, operation string, duration time.Duration) {
	m.DatabaseQueryDuration.WithLabelValues(databaseType, operation).Observe(duration.Seconds())
}

// RecordDatabaseError records a database error
func (m *Metrics) RecordDatabaseError(databaseType, errorType string) {
	m.DatabaseErrorsTotal.WithLabelValues(databaseType, errorType).Inc()
}

// SetHealthStatus sets the health status of a service
func (m *Metrics) SetHealthStatus(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.HealthStatus.WithLabelValues(service).Set(value)
}

// SetActiveConnections sets the number of active connections
func (m *Metrics) SetActiveConnections(count int) {
	m.ActiveConnections.Set(float64(count))
}
