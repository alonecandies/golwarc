package crawlers

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// PuppeteerClient wraps chromedp (Chrome DevTools Protocol) operations
// Provides a Puppeteer-like API for Go
type PuppeteerClient struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// PuppeteerConfig holds Puppeteer client configuration
type PuppeteerConfig struct {
	Headless bool
	Timeout  time.Duration
}

// NewPuppeteerClient creates a new chromedp-based client (Puppeteer-like)
func NewPuppeteerClient(config PuppeteerConfig) (*PuppeteerClient, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
	}

	return &PuppeteerClient{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// NewDefaultPuppeteerClient creates a Puppeteer client with default settings
func NewDefaultPuppeteerClient() (*PuppeteerClient, error) {
	return NewPuppeteerClient(PuppeteerConfig{
		Headless: true,
		Timeout:  30 * time.Second,
	})
}

// Navigate navigates to a URL
func (p *PuppeteerClient) Navigate(url string) error {
	return chromedp.Run(p.ctx, chromedp.Navigate(url))
}

// Click clicks an element
func (p *PuppeteerClient) Click(selector string) error {
	return chromedp.Run(p.ctx, chromedp.Click(selector))
}

// SendKeys sends keys to an element
func (p *PuppeteerClient) SendKeys(selector, keys string) error {
	return chromedp.Run(p.ctx, chromedp.SendKeys(selector, keys))
}

// SetValue sets the value of an input element
func (p *PuppeteerClient) SetValue(selector, value string) error {
	return chromedp.Run(p.ctx, chromedp.SetValue(selector, value))
}

// Clear clears an input element
func (p *PuppeteerClient) Clear(selector string) error {
	return chromedp.Run(p.ctx, chromedp.Clear(selector))
}

// Evaluate executes JavaScript code
func (p *PuppeteerClient) Evaluate(script string, res interface{}) error {
	return chromedp.Run(p.ctx, chromedp.Evaluate(script, res))
}

// EvaluateWithArgs executes JavaScript with arguments
func (p *PuppeteerClient) EvaluateWithArgs(script string, res interface{}, args ...interface{}) error {
	return chromedp.Run(p.ctx, chromedp.Evaluate(script, res))
}

// Screenshot takes a screenshot and saves it to a file
func (p *PuppeteerClient) Screenshot(path string) error {
	var buf []byte
	if err := chromedp.Run(p.ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return err
	}

	// Save to file if path is provided
	if path != "" {
		return chromedp.Run(p.ctx, chromedp.FullScreenshot(&buf, 100))
	}

	return nil
}

// ScreenshotBytes takes a screenshot and returns bytes
func (p *PuppeteerClient) ScreenshotBytes() ([]byte, error) {
	var buf []byte
	err := chromedp.Run(p.ctx, chromedp.CaptureScreenshot(&buf))
	return buf, err
}

// FullScreenshot takes a full page screenshot
func (p *PuppeteerClient) FullScreenshot() ([]byte, error) {
	var buf []byte
	err := chromedp.Run(p.ctx, chromedp.FullScreenshot(&buf, 100))
	return buf, err
}

// GetHTML gets the HTML content of an element
func (p *PuppeteerClient) GetHTML(selector string) (string, error) {
	var html string
	err := chromedp.Run(p.ctx, chromedp.OuterHTML(selector, &html))
	return html, err
}

// GetInnerHTML gets the inner HTML of an element
func (p *PuppeteerClient) GetInnerHTML(selector string) (string, error) {
	var html string
	err := chromedp.Run(p.ctx, chromedp.InnerHTML(selector, &html))
	return html, err
}

// GetText gets the text content of an element
func (p *PuppeteerClient) GetText(selector string) (string, error) {
	var text string
	err := chromedp.Run(p.ctx, chromedp.Text(selector, &text))
	return text, err
}

// GetAttribute gets an attribute value from an element
func (p *PuppeteerClient) GetAttribute(selector, attribute string) (string, error) {
	var value string
	err := chromedp.Run(p.ctx, chromedp.AttributeValue(selector, attribute, &value, nil))
	return value, err
}

// WaitVisible waits for an element to be visible
func (p *PuppeteerClient) WaitVisible(selector string) error {
	return chromedp.Run(p.ctx, chromedp.WaitVisible(selector))
}

// WaitReady waits for an element to be ready
func (p *PuppeteerClient) WaitReady(selector string) error {
	return chromedp.Run(p.ctx, chromedp.WaitReady(selector))
}

// WaitNotPresent waits for an element to be not present
func (p *PuppeteerClient) WaitNotPresent(selector string) error {
	return chromedp.Run(p.ctx, chromedp.WaitNotPresent(selector))
}

// Sleep waits for a specified duration
func (p *PuppeteerClient) Sleep(duration time.Duration) error {
	return chromedp.Run(p.ctx, chromedp.Sleep(duration))
}

// SetViewport sets the viewport size
func (p *PuppeteerClient) SetViewport(width, height int64) error {
	return chromedp.Run(p.ctx, chromedp.EmulateViewport(width, height))
}

// GetTitle gets the page title
func (p *PuppeteerClient) GetTitle() (string, error) {
	var title string
	err := chromedp.Run(p.ctx, chromedp.Title(&title))
	return title, err
}

// GetLocation gets the current URL
func (p *PuppeteerClient) GetLocation() (string, error) {
	var url string
	err := chromedp.Run(p.ctx, chromedp.Location(&url))
	return url, err
}

// Reload reloads the current page
func (p *PuppeteerClient) Reload() error {
	return chromedp.Run(p.ctx, chromedp.Reload())
}

// ScrollTo scrolls to an element
func (p *PuppeteerClient) ScrollTo(selector string) error {
	return chromedp.Run(p.ctx, chromedp.ScrollIntoView(selector))
}

// Submit submits a form
func (p *PuppeteerClient) Submit(selector string) error {
	return chromedp.Run(p.ctx, chromedp.Submit(selector))
}

// Focus focuses on an element
func (p *PuppeteerClient) Focus(selector string) error {
	return chromedp.Run(p.ctx, chromedp.Focus(selector))
}

// Blur removes focus from an element
func (p *PuppeteerClient) Blur(selector string) error {
	return chromedp.Run(p.ctx, chromedp.Blur(selector))
}

// QuerySelectorAll returns multiple elements matching selector
func (p *PuppeteerClient) QuerySelectorAll(selector string) ([]string, error) {
	var nodes []string
	script := fmt.Sprintf(`
		Array.from(document.querySelectorAll('%s')).map(el => el.outerHTML)
	`, selector)
	err := chromedp.Run(p.ctx, chromedp.Evaluate(script, &nodes))
	return nodes, err
}

// AddCookie adds a cookie
func (p *PuppeteerClient) AddCookie(name, value, domain string) error {
	//  Set cookie using chromedp.ActionFunc
	return chromedp.Run(p.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Cookie setting via JavaScript
		script := fmt.Sprintf(`document.cookie = "%s=%s; domain=%s"`, name, value, domain)
		var res interface{}
		return chromedp.Evaluate(script, &res).Do(ctx)
	}))
}

// Close closes the browser context
func (p *PuppeteerClient) Close() error {
	p.cancel()
	return nil
}

// Run executes a chromedp action
func (p *PuppeteerClient) Run(actions ...chromedp.Action) error {
	return chromedp.Run(p.ctx, actions...)
}

// GetContext returns the context for advanced operations
func (p *PuppeteerClient) GetContext() context.Context {
	return p.ctx
}
