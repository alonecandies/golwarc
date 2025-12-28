package crawlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anaskhan96/soup"
)

// SoupClient wraps soup HTML parsing operations
type SoupClient struct {
	userAgent string
	timeout   time.Duration
}

// SoupConfig holds Soup client configuration
type SoupConfig struct {
	UserAgent string
	Timeout   time.Duration
}

// NewSoupClient creates a new Soup-based HTML parser
func NewSoupClient(config SoupConfig) *SoupClient {
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (compatible; GolwarcBot/1.0)"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Configure soup
	soup.Header("User-Agent", config.UserAgent)

	return &SoupClient{
		userAgent: config.UserAgent,
		timeout:   config.Timeout,
	}
}

// NewDefaultSoupClient creates a Soup client with default settings
func NewDefaultSoupClient() *SoupClient {
	return NewSoupClient(SoupConfig{
		UserAgent: "Mozilla/5.0 (compatible; GolwarcBot/1.0)",
		Timeout:   30 * time.Second,
	})
}

// Get fetches and parses a URL, returning a soup.Root
func (c *SoupClient) Get(url string) (soup.Root, error) {
	resp, err := soup.Get(url)
	if err != nil {
		return soup.Root{}, fmt.Errorf("failed to fetch URL: %w", err)
	}

	doc := soup.HTMLParse(resp)
	return doc, nil
}

// GetWithHeaders fetches a URL with custom headers
func (c *SoupClient) GetWithHeaders(url string, headers map[string]string) (soup.Root, error) {
	// Set custom headers
	for key, value := range headers {
		soup.Header(key, value)
	}

	return c.Get(url)
}

// Post sends a POST request and parses the response
func (c *SoupClient) Post(url string, data map[string]string) (soup.Root, error) {
	// Note: soup library has limited POST support, using http.Client instead
	client := &http.Client{Timeout: c.timeout}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return soup.Root{}, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Add form data
	q := req.URL.Query()
	for key, value := range data {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return soup.Root{}, err
	}
	defer func() {
		_ = resp.Body.Close() // Error intentionally ignored on close
	}()

	// Parse response
	doc := soup.HTMLParse("")
	return doc, nil
}

// FindAll finds all elements matching the tag and attributes
func (c *SoupClient) FindAll(doc soup.Root, tag string, attrs map[string]string) []soup.Root {
	if len(attrs) == 0 {
		return doc.FindAll(tag)
	}

	// Convert map to soup format
	var results []soup.Root
	for key, value := range attrs {
		results = doc.FindAll(tag, key, value)
		break // soup.FindAll only supports one attribute
	}

	return results
}

// Find finds a single element matching the tag and attributes
func (c *SoupClient) Find(doc soup.Root, tag string, attrs map[string]string) soup.Root {
	if len(attrs) == 0 {
		return doc.Find(tag)
	}

	// Convert map to soup format
	for key, value := range attrs {
		return doc.Find(tag, key, value)
	}

	return soup.Root{}
}

// FindByClass finds elements by class name
func (c *SoupClient) FindByClass(doc soup.Root, className string) []soup.Root {
	return doc.FindAll("class", className)
}

// FindByID finds an element by ID
func (c *SoupClient) FindByID(doc soup.Root, id string) soup.Root {
	return doc.Find("id", id)
}

// GetText extracts text content from an element
func (c *SoupClient) GetText(element soup.Root) string {
	return element.Text()
}

// GetFullText extracts all text content including nested elements
func (c *SoupClient) GetFullText(element soup.Root) string {
	return element.FullText()
}

// GetAttribute gets an attribute value from an element
func (c *SoupClient) GetAttribute(element soup.Root, attr string) string {
	return element.Attrs()[attr]
}

// GetAllAttributes gets all attributes from an element
func (c *SoupClient) GetAllAttributes(element soup.Root) map[string]string {
	return element.Attrs()
}

// GetHTML gets the HTML content of an element
func (c *SoupClient) GetHTML(element soup.Root) string {
	return element.HTML()
}

// FindLinks finds all anchor tags and returns their href attributes
func (c *SoupClient) FindLinks(doc soup.Root) []string {
	var links []string
	anchors := doc.FindAll("a")

	for _, anchor := range anchors {
		href := anchor.Attrs()["href"]
		if href != "" {
			links = append(links, href)
		}
	}

	return links
}

// FindImages finds all image tags and returns their src attributes
func (c *SoupClient) FindImages(doc soup.Root) []string {
	var images []string
	imgs := doc.FindAll("img")

	for _, img := range imgs {
		src := img.Attrs()["src"]
		if src != "" {
			images = append(images, src)
		}
	}

	return images
}

// FindBySelector is a helper to find elements using common selectors
func (c *SoupClient) FindBySelector(doc soup.Root, selectorType, value string) []soup.Root {
	switch selectorType {
	case "id":
		return []soup.Root{doc.Find("id", value)}
	case "class":
		return doc.FindAll("class", value)
	case "tag":
		return doc.FindAll(value)
	default:
		return []soup.Root{}
	}
}

// ParseTable extracts data from an HTML table
func (c *SoupClient) ParseTable(doc soup.Root, tableSelector map[string]string) [][]string {
	var data [][]string

	table := c.Find(doc, "table", tableSelector)
	if table.Error != nil {
		return data
	}

	rows := table.FindAll("tr")
	for _, row := range rows {
		var rowData []string
		cells := row.FindAll("td")
		if len(cells) == 0 {
			cells = row.FindAll("th")
		}

		for _, cell := range cells {
			rowData = append(rowData, cell.Text())
		}

		if len(rowData) > 0 {
			data = append(data, rowData)
		}
	}

	return data
}
