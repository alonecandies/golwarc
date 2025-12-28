package crawlers

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// ValidateURL validates a URL for crawling
// Returns an error if the URL is invalid or potentially dangerous
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Require absolute URL with scheme
	if parsed.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http or https)")
	}

	// Only allow http and https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http and https schemes are allowed, got: %s", parsed.Scheme)
	}

	// Require a host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	// Block localhost and private IPs for security (SSRF protection)
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("localhost URLs are not allowed")
	}

	// Block private IP ranges (basic check)
	if strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.") {
		return fmt.Errorf("private IP addresses are not allowed")
	}

	return nil
}

// CollyClient wraps Colly crawler operations
type CollyClient struct {
	collector *colly.Collector
}

// CollyConfig holds Colly crawler configuration
type CollyConfig struct {
	UserAgent      string
	AllowedDomains []string
	MaxDepth       int
	Async          bool
	Parallelism    int
	Delay          time.Duration
}

// NewCollyClient creates a new Colly-based crawler
func NewCollyClient(config CollyConfig) *CollyClient {
	c := colly.NewCollector(
		colly.UserAgent(config.UserAgent),
		colly.AllowedDomains(config.AllowedDomains...),
		colly.MaxDepth(config.MaxDepth),
		colly.Async(config.Async),
	)

	// Set parallelism and delay
	if config.Parallelism > 0 {
		if err := c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: config.Parallelism,
			Delay:       config.Delay,
		}); err != nil {
			// Log warning but continue - limit rule failure is non-fatal
			fmt.Printf("warning: failed to set limit rule: %v\n", err)
		}
	}

	return &CollyClient{
		collector: c,
	}
}

// NewDefaultCollyClient creates a Colly client with default settings
func NewDefaultCollyClient() *CollyClient {
	return NewCollyClient(CollyConfig{
		UserAgent:   "Mozilla/5.0 (compatible; GolwarcBot/1.0)",
		MaxDepth:    3,
		Async:       false,
		Parallelism: 2,
		Delay:       1 * time.Second,
	})
}

// SetUserAgent sets the user agent
func (c *CollyClient) SetUserAgent(ua string) {
	c.collector.UserAgent = ua
}

// SetAllowedDomains sets the allowed domains
func (c *CollyClient) SetAllowedDomains(domains ...string) {
	c.collector.AllowedDomains = domains
}

// SetMaxDepth sets the maximum crawling depth
func (c *CollyClient) SetMaxDepth(depth int) {
	c.collector.MaxDepth = depth
}

// OnHTML registers a callback for HTML elements matching the selector
func (c *CollyClient) OnHTML(selector string, handler func(e *colly.HTMLElement)) {
	c.collector.OnHTML(selector, handler)
}

// OnRequest registers a callback before a request is made
func (c *CollyClient) OnRequest(handler func(r *colly.Request)) {
	c.collector.OnRequest(handler)
}

// OnResponse registers a callback after a response is received
func (c *CollyClient) OnResponse(handler func(r *colly.Response)) {
	c.collector.OnResponse(handler)
}

// OnError registers a callback when an error occurs
func (c *CollyClient) OnError(handler func(r *colly.Response, err error)) {
	c.collector.OnError(handler)
}

// OnScraped registers a callback after a page is scraped
func (c *CollyClient) OnScraped(handler func(r *colly.Response)) {
	c.collector.OnScraped(handler)
}

// Visit starts crawling from the given URL
func (c *CollyClient) Visit(url string) error {
	return c.collector.Visit(url)
}

// VisitMultiple visits multiple URLs
func (c *CollyClient) VisitMultiple(urls []string) error {
	for _, url := range urls {
		if err := c.collector.Visit(url); err != nil {
			return fmt.Errorf("failed to visit %s: %w", url, err)
		}
	}
	return nil
}

// Wait waits for all async requests to complete
func (c *CollyClient) Wait() {
	c.collector.Wait()
}

// WithCache enables cache for the collector (requires external storage implementation)
func (c *CollyClient) WithCache() error {
	// Note: Colly's storage must be implemented separately
	// Users can implement their own storage by implementing colly.Storage interface
	return nil
}

// Clone creates a new collector with the same configuration
func (c *CollyClient) Clone() *CollyClient {
	return &CollyClient{
		collector: c.collector.Clone(),
	}
}

// GetCollector returns the underlying Colly collector for advanced operations
func (c *CollyClient) GetCollector() *colly.Collector {
	return c.collector
}

// SetCookies sets cookies for requests
func (c *CollyClient) SetCookies(url string, cookies map[string]string) error {
	// Build cookie string from all cookies
	var cookieParts []string
	for name, value := range cookies {
		cookieParts = append(cookieParts, name+"="+value)
	}
	cookieString := strings.Join(cookieParts, "; ")

	// Register a single callback that sets all cookies
	c.collector.OnRequest(func(r *colly.Request) {
		if r.URL.String() == url || url == "" {
			r.Headers.Set("Cookie", cookieString)
		}
	})
	return nil
}

// SetHeaders sets custom headers for requests
func (c *CollyClient) SetHeaders(headers map[string]string) {
	c.collector.OnRequest(func(r *colly.Request) {
		for key, value := range headers {
			r.Headers.Set(key, value)
		}
	})
}
