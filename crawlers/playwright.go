package crawlers

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// PlaywrightClient wraps Playwright browser automation
type PlaywrightClient struct {
	pw        *playwright.Playwright
	browser   playwright.Browser
	page      playwright.Page
	ctx       context.Context
	rateLimit time.Duration
}

// PlaywrightConfig holds Playwright configuration
type PlaywrightConfig struct {
	BrowserType string // "chromium", "firefox", "webkit"
	Headless    bool
	Timeout     time.Duration
	RateLimit   time.Duration // Delay between navigation calls
}

// NewPlaywrightClient creates a new Playwright client
func NewPlaywrightClient(config PlaywrightConfig) (*PlaywrightClient, error) {
	if config.BrowserType == "" {
		config.BrowserType = "chromium"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start Playwright: %w", err)
	}

	var browser playwright.Browser

	browserOpts := playwright.BrowserTypeLaunchOptions{
		Headless: &config.Headless,
	}

	switch config.BrowserType {
	case "chromium":
		browser, err = pw.Chromium.Launch(browserOpts)
	case "firefox":
		browser, err = pw.Firefox.Launch(browserOpts)
	case "webkit":
		browser, err = pw.WebKit.Launch(browserOpts)
	default:
		_ = pw.Stop() // Best effort cleanup
		return nil, fmt.Errorf("unsupported browser type: %s", config.BrowserType)
	}

	if err != nil {
		_ = pw.Stop() // Best effort cleanup
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		_ = browser.Close() // Best effort cleanup
		_ = pw.Stop()       // Best effort cleanup
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	page.SetDefaultTimeout(float64(config.Timeout.Milliseconds()))

	return &PlaywrightClient{
		pw:        pw,
		browser:   browser,
		page:      page,
		ctx:       context.Background(),
		rateLimit: config.RateLimit,
	}, nil
}

// NewPage creates a new page
func (p *PlaywrightClient) NewPage() (playwright.Page, error) {
	return p.browser.NewPage()
}

// Navigate navigates to a URL with rate limiting
func (p *PlaywrightClient) Navigate(url string) error {
	// Apply rate limiting if configured
	if p.rateLimit > 0 {
		time.Sleep(p.rateLimit)
	}
	_, err := p.page.Goto(url)
	return err
}

// Click clicks an element using locator-based API
func (p *PlaywrightClient) Click(selector string) error {
	return p.page.Locator(selector).Click()
}

// Fill fills an input field using locator-based API
func (p *PlaywrightClient) Fill(selector, value string) error {
	return p.page.Locator(selector).Fill(value)
}

// Type types text into an element using locator-based API
func (p *PlaywrightClient) Type(selector, text string) error {
	return p.page.Locator(selector).PressSequentially(text)
}

// Press presses a key using locator-based API
func (p *PlaywrightClient) Press(selector, key string) error {
	return p.page.Locator(selector).Press(key)
}

// Evaluate executes JavaScript code
func (p *PlaywrightClient) Evaluate(script string) (interface{}, error) {
	return p.page.Evaluate(script)
}

// EvaluateHandle executes JavaScript and returns a handle
func (p *PlaywrightClient) EvaluateHandle(script string) (playwright.JSHandle, error) {
	return p.page.EvaluateHandle(script)
}

// Screenshot takes a screenshot
func (p *PlaywrightClient) Screenshot(path string) error {
	_, err := p.page.Screenshot(playwright.PageScreenshotOptions{
		Path: &path,
	})
	return err
}

// ScreenshotBytes takes a screenshot and returns bytes
func (p *PlaywrightClient) ScreenshotBytes() ([]byte, error) {
	return p.page.Screenshot()
}

// GetContent gets the HTML content of the page
func (p *PlaywrightClient) GetContent() (string, error) {
	return p.page.Content()
}

// GetTitle gets the page title
func (p *PlaywrightClient) GetTitle() (string, error) {
	return p.page.Title()
}

// GetURL gets the current URL
func (p *PlaywrightClient) GetURL() string {
	return p.page.URL()
}

// WaitForSelector waits for an element to appear using locator-based API
func (p *PlaywrightClient) WaitForSelector(selector string) error {
	return p.page.Locator(selector).WaitFor()
}

// WaitForTimeout waits for a specified duration
// Note: Using time.Sleep instead of deprecated page.WaitForTimeout
func (p *PlaywrightClient) WaitForTimeout(timeout time.Duration) {
	time.Sleep(timeout)
}

// WaitForLoadState waits for the page to reach a specific load state
func (p *PlaywrightClient) WaitForLoadState(state string) error {
	loadState := playwright.LoadState(state)
	return p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: &loadState,
	})
}

// QuerySelector returns a Locator for a single element.
//
// Deprecated: Use Locator() directly for better performance and reliability.
func (p *PlaywrightClient) QuerySelector(selector string) playwright.Locator {
	return p.page.Locator(selector)
}

// QuerySelectorAll returns a Locator that can iterate over multiple elements.
//
// Deprecated: Use Locator() directly for better performance and reliability.
func (p *PlaywrightClient) QuerySelectorAll(selector string) playwright.Locator {
	return p.page.Locator(selector)
}

// Locator returns a Playwright Locator for the given selector
// This is the recommended way to interact with elements
func (p *PlaywrightClient) Locator(selector string) playwright.Locator {
	return p.page.Locator(selector)
}

// GetText gets text content of an element using locator-based API
func (p *PlaywrightClient) GetText(selector string) (string, error) {
	return p.page.Locator(selector).TextContent()
}

// GetAttribute gets an attribute value using locator-based API
func (p *PlaywrightClient) GetAttribute(selector, name string) (string, error) {
	return p.page.Locator(selector).GetAttribute(name)
}

// SetViewportSize sets the viewport size
func (p *PlaywrightClient) SetViewportSize(width, height int) error {
	return p.page.SetViewportSize(width, height)
}

// GoBack navigates back
func (p *PlaywrightClient) GoBack() error {
	_, err := p.page.GoBack()
	return err
}

// GoForward navigates forward
func (p *PlaywrightClient) GoForward() error {
	_, err := p.page.GoForward()
	return err
}

// Reload reloads the page
func (p *PlaywrightClient) Reload() error {
	_, err := p.page.Reload()
	return err
}

// SetExtraHTTPHeaders sets extra HTTP headers
func (p *PlaywrightClient) SetExtraHTTPHeaders(headers map[string]string) error {
	return p.page.SetExtraHTTPHeaders(headers)
}

// AddCookie adds a cookie
func (p *PlaywrightClient) AddCookie(name, value, domain string) error {
	cookie := playwright.OptionalCookie{
		Name:   name,
		Value:  value,
		Domain: &domain,
	}
	if len(p.browser.Contexts()) > 0 {
		return p.browser.Contexts()[0].AddCookies([]playwright.OptionalCookie{cookie})
	}
	return fmt.Errorf("no browser context available")
}

// GetCookies gets all cookies
func (p *PlaywrightClient) GetCookies() ([]playwright.Cookie, error) {
	if len(p.browser.Contexts()) > 0 {
		return p.browser.Contexts()[0].Cookies()
	}
	return nil, fmt.Errorf("no browser context available")
}

// Close closes the browser and Playwright
func (p *PlaywrightClient) Close() error {
	if err := p.page.Close(); err != nil {
		return err
	}

	if err := p.browser.Close(); err != nil {
		return err
	}

	if err := p.pw.Stop(); err != nil {
		return err
	}

	return nil
}

// GetPage returns the current page for advanced operations
func (p *PlaywrightClient) GetPage() playwright.Page {
	return p.page
}

// GetBrowser returns the browser instance
func (p *PlaywrightClient) GetBrowser() playwright.Browser {
	return p.browser
}
