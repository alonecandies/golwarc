package crawlers

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

// SeleniumClient wraps Selenium WebDriver operations
type SeleniumClient struct {
	driver  selenium.WebDriver
	service *selenium.Service
}

// SeleniumConfig holds Selenium configuration
type SeleniumConfig struct {
	BrowserName string // "chrome", "firefox", etc.
	DriverPath  string
	Port        int
	Headless    bool
	RemoteURL   string // Optional: use remote Selenium server
}

// NewSeleniumClient creates a new Selenium WebDriver client
func NewSeleniumClient(config SeleniumConfig) (*SeleniumClient, error) {
	var driver selenium.WebDriver
	var service *selenium.Service
	var err error

	// Use remote URL if provided, otherwise start local service
	if config.RemoteURL != "" {
		caps := selenium.Capabilities{"browserName": config.BrowserName}
		if config.Headless && config.BrowserName == "chrome" {
			chromeArgs := map[string]interface{}{
				"args": []string{"--headless", "--no-sandbox", "--disable-dev-shm-usage"},
			}
			caps["goog:chromeOptions"] = chromeArgs
		}

		driver, err = selenium.NewRemote(caps, config.RemoteURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote driver: %w", err)
		}
	} else {
		// Start local service
		if config.Port == 0 {
			config.Port = 4444
		}

		opts := []selenium.ServiceOption{}
		service, err = selenium.NewSeleniumService(config.DriverPath, config.Port, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to start Selenium service: %w", err)
		}

		caps := selenium.Capabilities{"browserName": config.BrowserName}
		if config.Headless && config.BrowserName == "chrome" {
			chromeArgs := map[string]interface{}{
				"args": []string{"--headless", "--no-sandbox", "--disable-dev-shm-usage"},
			}
			caps["goog:chromeOptions"] = chromeArgs
		}

		driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", config.Port))
		if err != nil {
			_ = service.Stop() // Best effort cleanup
			return nil, fmt.Errorf("failed to create driver: %w", err)
		}
	}

	return &SeleniumClient{
		driver:  driver,
		service: service,
	}, nil
}

// Navigate navigates to a URL
func (s *SeleniumClient) Navigate(url string) error {
	return s.driver.Get(url)
}

// FindElement finds an element by selector
func (s *SeleniumClient) FindElement(by, value string) (selenium.WebElement, error) {
	return s.driver.FindElement(by, value)
}

// FindElements finds multiple elements by selector
func (s *SeleniumClient) FindElements(by, value string) ([]selenium.WebElement, error) {
	return s.driver.FindElements(by, value)
}

// Click clicks an element
func (s *SeleniumClient) Click(element selenium.WebElement) error {
	return element.Click()
}

// SendKeys sends keys to an element
func (s *SeleniumClient) SendKeys(element selenium.WebElement, keys string) error {
	return element.SendKeys(keys)
}

// Clear clears an input element
func (s *SeleniumClient) Clear(element selenium.WebElement) error {
	return element.Clear()
}

// GetText gets the text content of an element
func (s *SeleniumClient) GetText(element selenium.WebElement) (string, error) {
	return element.Text()
}

// GetAttribute gets an attribute value from an element
func (s *SeleniumClient) GetAttribute(element selenium.WebElement, name string) (string, error) {
	return element.GetAttribute(name)
}

// ExecuteScript executes JavaScript code
func (s *SeleniumClient) ExecuteScript(script string, args []interface{}) (interface{}, error) {
	return s.driver.ExecuteScript(script, args)
}

// Screenshot takes a screenshot and returns the image data
func (s *SeleniumClient) Screenshot() ([]byte, error) {
	return s.driver.Screenshot()
}

// GetPageSource gets the HTML source of the current page
func (s *SeleniumClient) GetPageSource() (string, error) {
	return s.driver.PageSource()
}

// GetTitle gets the page title
func (s *SeleniumClient) GetTitle() (string, error) {
	return s.driver.Title()
}

// GetCurrentURL gets the current URL
func (s *SeleniumClient) GetCurrentURL() (string, error) {
	return s.driver.CurrentURL()
}

// Back navigates back
func (s *SeleniumClient) Back() error {
	return s.driver.Back()
}

// Forward navigates forward
func (s *SeleniumClient) Forward() error {
	return s.driver.Forward()
}

// Refresh refreshes the page
func (s *SeleniumClient) Refresh() error {
	return s.driver.Refresh()
}

// WaitForElement waits for an element to be present
func (s *SeleniumClient) WaitForElement(by, value string, timeout time.Duration) (selenium.WebElement, error) {
	endTime := time.Now().Add(timeout)
	for time.Now().Before(endTime) {
		element, err := s.driver.FindElement(by, value)
		if err == nil {
			return element, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for element")
}

// SwitchToFrame switches to an iframe
func (s *SeleniumClient) SwitchToFrame(frame interface{}) error {
	return s.driver.SwitchFrame(frame)
}

// SwitchToDefaultContent switches back to the main document
func (s *SeleniumClient) SwitchToDefaultContent() error {
	return s.driver.SwitchFrame(nil)
}

// GetCookies gets all cookies
func (s *SeleniumClient) GetCookies() ([]selenium.Cookie, error) {
	return s.driver.GetCookies()
}

// AddCookie adds a cookie
func (s *SeleniumClient) AddCookie(cookie *selenium.Cookie) error {
	return s.driver.AddCookie(cookie)
}

// DeleteAllCookies deletes all cookies
func (s *SeleniumClient) DeleteAllCookies() error {
	return s.driver.DeleteAllCookies()
}

// SetWindowSize sets the window size
func (s *SeleniumClient) SetWindowSize(width, height int) error {
	return s.driver.ResizeWindow("", width, height)
}

// MaximizeWindow maximizes the window
func (s *SeleniumClient) MaximizeWindow() error {
	return s.driver.MaximizeWindow("")
}

// Close closes the current window
func (s *SeleniumClient) Close() error {
	if err := s.driver.Quit(); err != nil {
		return err
	}

	if s.service != nil {
		if err := s.service.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// GetDriver returns the underlying WebDriver for advanced operations
func (s *SeleniumClient) GetDriver() selenium.WebDriver {
	return s.driver
}
