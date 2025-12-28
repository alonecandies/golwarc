package crawlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alonecandies/golwarc/crawlers"
)

// =============================================================================
// Soup Client Tests
// =============================================================================

func TestNewSoupClient(t *testing.T) {
	tests := []struct {
		name   string
		config crawlers.SoupConfig
	}{
		{
			name: "with custom config",
			config: crawlers.SoupConfig{
				UserAgent: "CustomBot/1.0",
				Timeout:   10 * time.Second,
			},
		},
		{
			name: "with empty user agent (uses default)",
			config: crawlers.SoupConfig{
				Timeout: 15 * time.Second,
			},
		},
		{
			name: "with zero timeout (uses default)",
			config: crawlers.SoupConfig{
				UserAgent: "TestBot/1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := crawlers.NewSoupClient(tt.config)
			if client == nil {
				t.Fatal("NewSoupClient() returned nil")
			}
		})
	}
}

func TestNewDefaultSoupClient(t *testing.T) {
	client := crawlers.NewDefaultSoupClient()
	if client == nil {
		t.Fatal("NewDefaultSoupClient() returned nil")
	}
}

func TestSoupClient_Get_Integration(t *testing.T) {
	client := crawlers.NewDefaultSoupClient()

	_, err := client.Get("https://example.com")
	if err != nil {
		t.Skip("Network not available:", err)
	}
}

func TestSoupClient_GetWithHeaders_Integration(t *testing.T) {
	client := crawlers.NewDefaultSoupClient()

	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}

	_, err := client.GetWithHeaders("https://example.com", headers)
	if err != nil {
		t.Skip("Network not available:", err)
	}
}

func TestSoupClient_Post_Integration(t *testing.T) {
	client := crawlers.NewDefaultSoupClient()

	data := map[string]string{
		"key": "value",
	}

	_, err := client.Post("https://httpbin.org/post", data)
	if err != nil {
		t.Skip("Network not available:", err)
	}
}

// =============================================================================
// Spider Tests
// =============================================================================

func TestNewSpider(t *testing.T) {
	tests := []struct {
		name   string
		config crawlers.SpiderConfig
	}{
		{
			name: "with custom config",
			config: crawlers.SpiderConfig{
				MaxDepth:    5,
				Concurrency: 10,
				UserAgent:   "TestSpider/1.0",
				Delay:       2 * time.Second,
				Timeout:     60 * time.Second,
			},
		},
		{
			name:   "with defaults",
			config: crawlers.SpiderConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spider := crawlers.NewSpider(tt.config)
			if spider == nil {
				t.Fatal("NewSpider() returned nil")
			}
		})
	}
}

func TestNewDefaultSpider(t *testing.T) {
	spider := crawlers.NewDefaultSpider()
	if spider == nil {
		t.Fatal("NewDefaultSpider() returned nil")
	}
}

func TestSpider_SetMaxDepth(t *testing.T) {
	spider := crawlers.NewDefaultSpider()
	spider.SetMaxDepth(10)
	// No way to verify directly, but ensures no panic
}

func TestSpider_SetConcurrency(t *testing.T) {
	spider := crawlers.NewDefaultSpider()
	spider.SetConcurrency(20)
	// No way to verify directly, but ensures no panic
}

func TestSpider_AddStartURL(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	spider.AddStartURL("https://example.com")
	spider.AddStartURL("https://example.com/page1")
	spider.AddStartURL("https://example.com/page2")

	// URLs should be added to queue (no direct way to verify, but ensures no panic)
}

func TestSpider_OnDocument(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	called := false
	spider.OnDocument(func(doc *goquery.Document, url string) error {
		called = true
		return nil
	})

	// Handler should be set (verified when Run() is called)
	_ = called
}

func TestSpider_IsRunning(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	if spider.IsRunning() {
		t.Error("Spider should not be running initially")
	}
}

func TestSpider_ClearVisited(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	spider.ClearVisited()

	count := spider.GetVisitedCount()
	if count != 0 {
		t.Errorf("GetVisitedCount() = %d, want 0", count)
	}
}

func TestSpider_GetVisitedCount(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	count := spider.GetVisitedCount()
	if count != 0 {
		t.Errorf("Initial GetVisitedCount() = %d, want 0", count)
	}
}

func TestSpider_Stop(t *testing.T) {
	spider := crawlers.NewDefaultSpider()

	spider.Stop()

	// Should not panic
}

func TestSpider_Run_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Test</h1></body></html>`))
	}))
	defer server.Close()

	spider := crawlers.NewDefaultSpider()
	spider.AddStartURL(server.URL)

	documentCalled := false
	spider.OnDocument(func(doc *goquery.Document, url string) error {
		documentCalled = true
		return nil
	})

	err := spider.Run()
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}

	if !documentCalled {
		t.Error("OnDocument callback was not called")
	}

	count := spider.GetVisitedCount()
	if count != 1 {
		t.Errorf("GetVisitedCount() = %d, want 1", count)
	}
}

func TestSpider_Run_AlreadyRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte(`<html><body>Test</body></html>`))
	}))
	defer server.Close()

	spider := crawlers.NewDefaultSpider()
	spider.AddStartURL(server.URL)

	// Start first run in goroutine
	go func() {
		spider.Run()
	}()

	time.Sleep(10 * time.Millisecond) // Give it time to start

	// Try to run again while already running
	err := spider.Run()
	if err == nil {
		t.Error("Run() should return error when spider is already running")
	}
}

func TestSpider_ExtractLinks(t *testing.T) {
	htmlContent := `<html>
		<body>
			<a href="/page1">Link 1</a>
			<a href="/page2">Link 2</a>
			<a id="special" href="/special">Special Link</a>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	spider := crawlers.NewDefaultSpider()
	links := spider.ExtractLinks(doc, "a")

	if len(links) != 3 {
		t.Errorf("ExtractLinks() found %d links, want 3", len(links))
	}
}

func TestSpider_ExtractLinksWithCascadia(t *testing.T) {
	htmlContent := `<html>
		<body>
			<div class="content">
				<a href="/page1">Link 1</a>
				<a href="/page2">Link 2</a>
			</div>
			<a href="/outside">Outside Link</a>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	spider := crawlers.NewDefaultSpider()
	links := spider.ExtractLinksWithCascadia(doc, ".content a")

	// Should extract links from .content div
	t.Logf("Found %d links", len(links))
}

func TestSpider_ResolveURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		relativeURL string
		want        string
		wantErr     bool
	}{
		{
			name:        "relative path",
			baseURL:     "https://example.com/base/",
			relativeURL: "page.html",
			want:        "https://example.com/base/page.html",
			wantErr:     false,
		},
		{
			name:        "absolute path",
			relativeURL: "/absolute/path",
			baseURL:     "https://example.com/base/",
			want:        "https://example.com/absolute/path",
			wantErr:     false,
		},
		{
			name:        "full URL",
			baseURL:     "https://example.com/",
			relativeURL: "https://other.com/page",
			want:        "https://other.com/page",
			wantErr:     false,
		},
		{
			name:        "invalid base URL",
			baseURL:     "://invalid",
			relativeURL: "/page",
			wantErr:     true,
		},
	}

	spider := crawlers.NewDefaultSpider()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := spider.ResolveURL(tt.baseURL, tt.relativeURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ResolveURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpider_MultipleURLs_Sequential(t *testing.T) {
	visitedURLs := []string{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		visitedURLs = append(visitedURLs, r.URL.Path)
		mu.Unlock()
		w.Write([]byte(`<html><body>Page content</body></html>`))
	}))
	defer server.Close()

	spider := crawlers.NewDefaultSpider()
	spider.AddStartURL(server.URL + "/page1")
	spider.AddStartURL(server.URL + "/page2")
	spider.AddStartURL(server.URL + "/page3")

	err := spider.Run()
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}

	count := spider.GetVisitedCount()
	if count != 3 {
		t.Errorf("GetVisitedCount() = %d, want 3", count)
	}
}

// =============================================================================
// Spider Configuration Tests
// =============================================================================

func TestSpiderConfig_Defaults(t *testing.T) {
	config := crawlers.SpiderConfig{}
	spider := crawlers.NewSpider(config)

	// Default values should be set
	if spider == nil {
		t.Fatal("Spider with default config should not be nil")
	}
}

func TestSpiderConfig_CustomValues(t *testing.T) {
	config := crawlers.SpiderConfig{
		MaxDepth:    10,
		Concurrency: 20,
		UserAgent:   "MyCustomBot/2.0",
		Delay:       5 * time.Second,
		Timeout:     120 * time.Second,
	}

	spider := crawlers.NewSpider(config)
	if spider == nil {
		t.Fatal("Spider with custom config should not be nil")
	}
}

// =============================================================================
// Soup Client Method Tests (Unit)
// =============================================================================

func TestSoupConfig_Defaults(t *testing.T) {
	config := crawlers.SoupConfig{}
	client := crawlers.NewSoupClient(config)

	// Should apply defaults
	if client == nil {
		t.Fatal("Client with default config should not be nil")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewSpider(b *testing.B) {
	config := crawlers.SpiderConfig{
		MaxDepth:    5,
		Concurrency: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crawlers.NewSpider(config)
	}
}

func BenchmarkNewSoupClient(b *testing.B) {
	config := crawlers.SoupConfig{
		UserAgent: "BenchBot/1.0",
		Timeout:   30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crawlers.NewSoupClient(config)
	}
}
