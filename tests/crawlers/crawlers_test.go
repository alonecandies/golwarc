package crawlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alonecandies/golwarc/crawlers"
	"github.com/gocolly/colly/v2"
)

// =============================================================================
// URL Validation Tests (SSRF Protection)
// =============================================================================

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Invalid URLs
		{"empty URL", "", true},
		{"no scheme", "example.com", true},
		{"no host", "http://", true},
		{"file scheme", "file:///etc/passwd", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"ftp scheme", "ftp://example.com", true},
		{"data scheme", "data:text/html,<script>alert(1)</script>", true},

		// SSRF protection - localhost
		{"localhost", "http://localhost/admin", true},
		{"127.0.0.1", "http://127.0.0.1:8080/api", true},
		{"IPv6 localhost", "http://[::1]/admin", true},
		{"LOCALHOST uppercase", "http://LOCALHOST/admin", true},

		// SSRF protection - private IPs
		{"10.x.x.x", "http://10.0.0.1/internal", true},
		{"192.168.x.x", "http://192.168.1.1/router", true},
		{"172.16.x.x", "http://172.16.0.1/internal", true},
		{"172.31.x.x", "http://172.31.255.255/internal", true},
		{"172.20.x.x", "http://172.20.1.1/internal", true},

		// Valid URLs
		{"http example.com", "http://example.com", false},
		{"https example.com", "https://example.com", false},
		{"https with path", "https://example.com/path/to/page", false},
		{"https with query", "https://example.com?q=test", false},
		{"https with port", "https://example.com:8443/api", false},
		{"https with fragment", "https://example.com/page#section", false},
		{"subdomain", "https://api.example.com/v1/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crawlers.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// CollyClient Constructor Tests
// =============================================================================

func TestNewDefaultCollyClient(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()
	if client == nil {
		t.Fatal("Client should not be nil")
	}

	collector := client.GetCollector()
	if collector == nil {
		t.Fatal("Collector should not be nil")
	}
}

func TestNewCollyClient(t *testing.T) {
	config := crawlers.CollyConfig{
		AllowedDomains: []string{"example.com"},
		MaxDepth:       2,
		Async:          false,
		Parallelism:    2,
		Delay:          1 * time.Second,
		UserAgent:      "TestBot/1.0",
	}

	client := crawlers.NewCollyClient(config)
	if client == nil {
		t.Fatal("Client should not be nil")
	}

	collector := client.GetCollector()
	if collector == nil {
		t.Fatal("Collector should not be nil")
	}

	// Verify user agent was set
	if collector.UserAgent != "TestBot/1.0" {
		t.Errorf("UserAgent = %v, want TestBot/1.0", collector.UserAgent)
	}

	// Verify max depth
	if collector.MaxDepth != 2 {
		t.Errorf("MaxDepth = %v, want 2", collector.MaxDepth)
	}
}

func TestNewCollyClient_WithZeroParallelism(t *testing.T) {
	config := crawlers.CollyConfig{
		UserAgent:   "TestBot/1.0",
		Parallelism: 0, // Should not set limit rule
	}

	client := crawlers.NewCollyClient(config)
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// =============================================================================
// CollyClient Configuration Tests
// =============================================================================

func TestCollyClient_SetUserAgent(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	client.SetUserAgent("CustomBot/2.0")

	collector := client.GetCollector()
	if collector.UserAgent != "CustomBot/2.0" {
		t.Errorf("UserAgent = %v, want CustomBot/2.0", collector.UserAgent)
	}
}

func TestCollyClient_SetAllowedDomains(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	client.SetAllowedDomains("example.com", "test.com")

	collector := client.GetCollector()
	if len(collector.AllowedDomains) != 2 {
		t.Errorf("AllowedDomains length = %v, want 2", len(collector.AllowedDomains))
	}
}

func TestCollyClient_SetMaxDepth(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	client.SetMaxDepth(5)

	collector := client.GetCollector()
	if collector.MaxDepth != 5 {
		t.Errorf("MaxDepth = %v, want 5", collector.MaxDepth)
	}
}

// =============================================================================
// CollyClient Callback Tests with httptest
// =============================================================================

func TestCollyClient_OnRequest(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains() // Allow all domains for testing

	requestCalled := false
	var requestURL string

	client.OnRequest(func(r *colly.Request) {
		requestCalled = true
		requestURL = r.URL.String()
	})

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	if !requestCalled {
		t.Error("OnRequest callback was not called")
	}
	// Note: URL may have trailing slash added by server/colly
	if !strings.HasPrefix(requestURL, server.URL) {
		t.Errorf("Request URL = %v, want to start with %v", requestURL, server.URL)
	}
}

func TestCollyClient_OnResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test Response</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	responseCalled := false
	var statusCode int

	client.OnResponse(func(r *colly.Response) {
		responseCalled = true
		statusCode = r.StatusCode
	})

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	if !responseCalled {
		t.Error("OnResponse callback was not called")
	}
	if statusCode != 200 {
		t.Errorf("Status code = %v, want 200", statusCode)
	}
}

func TestCollyClient_OnHTML(t *testing.T) {
	htmlContent := `<html>
		<head><title>Test Page</title></head>
		<body>
			<h1>Hello World</h1>
			<a href="/page1">Link 1</a>
			<a href="/page2">Link 2</a>
		</body>
	</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	var foundTitle string
	linkCount := 0

	client.OnHTML("title", func(e *colly.HTMLElement) {
		foundTitle = e.Text
	})

	client.OnHTML("a", func(e *colly.HTMLElement) {
		linkCount++
	})

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	if foundTitle != "Test Page" {
		t.Errorf("Title = %v, want 'Test Page'", foundTitle)
	}
	if linkCount != 2 {
		t.Errorf("Link count = %v, want 2", linkCount)
	}
}

func TestCollyClient_OnError(t *testing.T) {
	// Server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	errorCalled := false
	var errorStatusCode int

	client.OnError(func(r *colly.Response, err error) {
		errorCalled = true
		if r != nil {
			errorStatusCode = r.StatusCode
		}
	})

	_ = client.Visit(server.URL)

	if !errorCalled {
		t.Error("OnError callback was not called for 404 response")
	}
	if errorStatusCode != 404 {
		t.Errorf("Error status code = %v, want 404", errorStatusCode)
	}
}

func TestCollyClient_OnScraped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Scraped</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	scrapedCalled := false

	client.OnScraped(func(r *colly.Response) {
		scrapedCalled = true
	})

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	if !scrapedCalled {
		t.Error("OnScraped callback was not called")
	}
}

// =============================================================================
// CollyClient Visit Tests
// =============================================================================

func TestCollyClient_Visit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Success</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	err := client.Visit(server.URL)
	if err != nil {
		t.Errorf("Visit() error = %v", err)
	}
}

func TestCollyClient_VisitMultiple(t *testing.T) {
	visitCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		visitCount++
		mu.Unlock()
		w.Write([]byte("<html><body>Page</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
	}

	err := client.VisitMultiple(urls)
	if err != nil {
		t.Errorf("VisitMultiple() error = %v", err)
	}

	if visitCount != 3 {
		t.Errorf("Visit count = %v, want 3", visitCount)
	}
}

func TestCollyClient_VisitMultiple_Error(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains("invalid-domain-that-does-not-exist.com")

	urls := []string{
		"http://invalid-domain-that-does-not-exist.com/page1",
	}

	err := client.VisitMultiple(urls)
	if err == nil {
		t.Error("VisitMultiple() should return error for invalid URL")
	}
}

// =============================================================================
// CollyClient Advanced Features Tests
// =============================================================================

func TestCollyClient_Clone(t *testing.T) {
	original := crawlers.NewDefaultCollyClient()
	original.SetMaxDepth(5)

	cloned := original.Clone()
	if cloned == nil {
		t.Fatal("Clone() should not return nil")
	}

	// Verify it's a different instance
	cloned.SetMaxDepth(10)

	// Original should still have depth 5
	if original.GetCollector().MaxDepth != 5 {
		t.Error("Original collector was modified by clone")
	}
}

func TestCollyClient_GetCollector(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	collector := client.GetCollector()
	if collector == nil {
		t.Fatal("GetCollector() should not return nil")
	}

	// Verify it's the actual colly.Collector
	collector.MaxDepth = 99
	if client.GetCollector().MaxDepth != 99 {
		t.Error("GetCollector() should return the same instance")
	}
}

func TestCollyClient_WithCache(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	err := client.WithCache()
	// WithCache returns nil (implementation pending)
	if err != nil {
		t.Errorf("WithCache() error = %v, want nil", err)
	}
}

func TestCollyClient_SetCookies(t *testing.T) {
	var receivedCookies string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCookies = r.Header.Get("Cookie")
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	cookies := map[string]string{
		"session": "abc123",
		"user":    "testuser",
	}

	// Use empty URL to apply cookies to all requests
	err := client.SetCookies("", cookies)
	if err != nil {
		t.Fatalf("SetCookies() error = %v", err)
	}

	err = client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Check that cookies were sent
	if receivedCookies == "" {
		t.Error("No cookies received")
	}
	if !strings.Contains(receivedCookies, "session=abc123") {
		t.Errorf("Session cookie not found in request, got: %s", receivedCookies)
	}
	if !strings.Contains(receivedCookies, "user=testuser") {
		t.Errorf("User cookie not found in request, got: %s", receivedCookies)
	}
}

func TestCollyClient_SetHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	headers := map[string]string{
		"X-Custom-Header": "test-value",
		"Authorization":   "Bearer token123",
	}

	client.SetHeaders(headers)

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Check that headers were sent
	if receivedHeaders.Get("X-Custom-Header") != "test-value" {
		t.Errorf("Custom header = %v, want test-value", receivedHeaders.Get("X-Custom-Header"))
	}
	if receivedHeaders.Get("Authorization") != "Bearer token123" {
		t.Errorf("Authorization header = %v, want Bearer token123", receivedHeaders.Get("Authorization"))
	}
}

func TestCollyClient_Wait(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	config := crawlers.CollyConfig{
		Async:       true,
		Parallelism: 2,
	}

	client := crawlers.NewCollyClient(config)
	client.SetAllowedDomains()

	// Start async visits
	client.Visit(server.URL + "/1")
	client.Visit(server.URL + "/2")

	// Wait should block until all requests complete
	client.Wait()
	// If we get here, Wait() worked correctly
}

// =============================================================================
// CollyClient Edge Cases
// =============================================================================

func TestCollyClient_EmptyCookies(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	err := client.SetCookies("http://example.com", map[string]string{})
	if err != nil {
		t.Errorf("SetCookies() with empty map error = %v", err)
	}
}

func TestCollyClient_EmptyHeaders(t *testing.T) {
	client := crawlers.NewDefaultCollyClient()

	// Should not panic
	client.SetHeaders(map[string]string{})
}

func TestCollyClient_MultipleCallbacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	count := 0

	// Register multiple OnRequest callbacks
	client.OnRequest(func(r *colly.Request) {
		count++
	})
	client.OnRequest(func(r *colly.Request) {
		count++
	})

	err := client.Visit(server.URL)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	if count != 2 {
		t.Errorf("Callback count = %v, want 2", count)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkValidateURL_Valid(b *testing.B) {
	url := "https://example.com/path/to/page"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crawlers.ValidateURL(url)
	}
}

func BenchmarkValidateURL_Invalid(b *testing.B) {
	url := "http://localhost/admin"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crawlers.ValidateURL(url)
	}
}

func BenchmarkCollyClient_Visit(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Benchmark</body></html>"))
	}))
	defer server.Close()

	client := crawlers.NewDefaultCollyClient()
	client.SetAllowedDomains()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Visit(fmt.Sprintf("%s?id=%d", server.URL, i))
	}
}
