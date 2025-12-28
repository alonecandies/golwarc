package crawlers

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

// Spider is a custom web crawler using goquery and cascadia
type Spider struct {
	httpClient  *http.Client
	maxDepth    int
	concurrency int
	visited     map[string]bool
	visitedMu   sync.RWMutex
	queue       []string
	queueMu     sync.RWMutex
	userAgent   string
	delay       time.Duration
	onDocument  func(doc *goquery.Document, url string) error
	running     bool
	wg          sync.WaitGroup
}

// SpiderConfig holds Spider configuration
type SpiderConfig struct {
	MaxDepth    int
	Concurrency int
	UserAgent   string
	Delay       time.Duration
	Timeout     time.Duration
}

// NewSpider creates a new Spider crawler
func NewSpider(config SpiderConfig) *Spider {
	if config.MaxDepth == 0 {
		config.MaxDepth = 3
	}
	if config.Concurrency == 0 {
		config.Concurrency = 5
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (compatible; GolwarcBot/1.0)"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Spider{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		maxDepth:    config.MaxDepth,
		concurrency: config.Concurrency,
		userAgent:   config.UserAgent,
		delay:       config.Delay,
		visited:     make(map[string]bool),
		queue:       []string{},
		running:     false,
	}
}

// NewDefaultSpider creates a Spider with default settings
func NewDefaultSpider() *Spider {
	return NewSpider(SpiderConfig{
		MaxDepth:    3,
		Concurrency: 5,
		UserAgent:   "Mozilla/5.0 (compatible; GolwarcBot/1.0)",
		Delay:       1 * time.Second,
		Timeout:     30 * time.Second,
	})
}

// SetMaxDepth sets the maximum crawl depth
func (s *Spider) SetMaxDepth(depth int) {
	s.maxDepth = depth
}

// SetConcurrency sets the number of concurrent requests
func (s *Spider) SetConcurrency(n int) {
	s.concurrency = n
}

// AddStartURL adds a starting URL to the queue
func (s *Spider) AddStartURL(url string) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	s.queue = append(s.queue, url)
}

// OnDocument registers a callback for processing documents
func (s *Spider) OnDocument(handler func(doc *goquery.Document, url string) error) {
	s.onDocument = handler
}

// Run starts the crawler
func (s *Spider) Run() error {
	if s.running {
		return fmt.Errorf("spider is already running")
	}

	s.running = true
	defer func() { s.running = false }()

	sem := make(chan struct{}, s.concurrency)

	for len(s.queue) > 0 {
		s.queueMu.Lock()
		if len(s.queue) == 0 {
			s.queueMu.Unlock()
			break
		}
		currentURL := s.queue[0]
		s.queue = s.queue[1:]
		s.queueMu.Unlock()

		// Check if already visited
		s.visitedMu.RLock()
		isVisited := s.visited[currentURL]
		s.visitedMu.RUnlock()

		if isVisited {
			continue
		}

		// Mark as visited
		s.visitedMu.Lock()
		s.visited[currentURL] = true
		s.visitedMu.Unlock()

		sem <- struct{}{}
		s.wg.Add(1)

		go func(url string) {
			defer func() {
				<-sem
				s.wg.Done()
			}()

			if err := s.crawlURL(url); err != nil {
				fmt.Printf("Error crawling %s: %v\n", url, err)
			}

			// Rate limiting
			if s.delay > 0 {
				time.Sleep(s.delay)
			}
		}(currentURL)
	}

	s.wg.Wait()
	return nil
}

// crawlURL fetches and processes a single URL
func (s *Spider) crawlURL(urlStr string) error {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close() // Error intentionally ignored on close
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Call the document handler
	if s.onDocument != nil {
		if err := s.onDocument(doc, urlStr); err != nil {
			return err
		}
	}

	return nil
}

// ExtractLinks extracts links from a document using a CSS selector
func (s *Spider) ExtractLinks(doc *goquery.Document, selector string) []string {
	var links []string

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if exists {
			links = append(links, href)
		}
	})

	return links
}

// ExtractLinksWithCascadia extracts links using cascadia selector
func (s *Spider) ExtractLinksWithCascadia(doc *goquery.Document, selectorStr string) []string {
	var links []string

	selector, err := cascadia.Parse(selectorStr)
	if err != nil {
		return links
	}

	// Use cascadia with goquery - access nodes via Find to avoid embedded field warning
	doc.Find("*").Each(func(i int, sel *goquery.Selection) {
		if len(sel.Nodes) > 0 {
			nodes := cascadia.QueryAll(sel.Nodes[0], selector)
			for _, node := range nodes {
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	})

	return links
}

// ResolveURL resolves a relative URL against a base URL
func (s *Spider) ResolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	relative, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(relative).String(), nil
}

// Stop stops the crawler
func (s *Spider) Stop() {
	s.running = false
}

// IsRunning checks if the spider is currently running
func (s *Spider) IsRunning() bool {
	return s.running
}

// ClearVisited clears the visited URLs map
func (s *Spider) ClearVisited() {
	s.visitedMu.Lock()
	defer s.visitedMu.Unlock()
	s.visited = make(map[string]bool)
}

// GetVisitedCount returns the number of visited URLs
func (s *Spider) GetVisitedCount() int {
	s.visitedMu.RLock()
	defer s.visitedMu.RUnlock()
	return len(s.visited)
}
