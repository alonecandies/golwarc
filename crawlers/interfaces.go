package crawlers

import "github.com/gocolly/colly/v2"

// CrawlerClient defines the interface for web crawling operations
// This enables mocking in tests and provides a consistent API across different crawler implementations
type CrawlerClient interface {
	// Visit starts crawling from the given URL
	Visit(url string) error

	// VisitMultiple visits multiple URLs
	VisitMultiple(urls []string) error

	// Wait waits for all async requests to complete
	Wait()

	// OnHTML registers a callback for HTML elements matching the selector
	OnHTML(selector string, handler func(e *colly.HTMLElement))

	// OnRequest registers a callback before a request is made
	OnRequest(handler func(r *colly.Request))

	// OnResponse registers a callback after a response is received
	OnResponse(handler func(r *colly.Response))

	// OnError registers a callback when an error occurs
	OnError(handler func(r *colly.Response, err error))

	// OnScraped registers a callback after a page is scraped
	OnScraped(handler func(r *colly.Response))

	// SetUserAgent sets the user agent
	SetUserAgent(ua string)

	// SetHeaders sets custom headers for requests
	SetHeaders(headers map[string]string)
}

// Ensure CollyClient implements the CrawlerClient interface
var _ CrawlerClient = (*CollyClient)(nil)
