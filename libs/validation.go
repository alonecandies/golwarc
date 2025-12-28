package libs

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Validator provides input validation functionality
type Validator struct {
	// Additional validator configuration can be added here
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateURL validates a URL and checks for SSRF vulnerabilities.
// It blocks:
// - file:// and javascript:// schemes
// - localhost and loopback addresses
// - private IP ranges (10.x.x.x, 192.168.x.x, 172.16-31.x.x)
// - link-local addresses (169.254.x.x)
//
// Returns an error if the URL is invalid or potentially dangerous.
func (v *Validator) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	scheme := strings.ToLower(parsedURL.Scheme)
	if err := v.validateScheme(scheme); err != nil {
		return err
	}

	// Check hostname for SSRF
	hostname := parsedURL.Hostname()
	if err := v.validateHostname(hostname); err != nil {
		return err
	}

	return nil
}

// validateScheme checks if the URL scheme is allowed
func (v *Validator) validateScheme(scheme string) error {
	allowedSchemes := map[string]bool{
		"http":  true,
		"https": true,
	}

	if !allowedSchemes[scheme] {
		return fmt.Errorf("disallowed URL scheme: %s (only http/https allowed)", scheme)
	}

	return nil
}

// validateHostname checks for SSRF vulnerabilities in the hostname
func (v *Validator) validateHostname(hostname string) error {
	if hostname == "" {
		return errors.New("hostname cannot be empty")
	}

	// Block localhost variations
	if isLocalhost(hostname) {
		return errors.New("localhost URLs are not allowed for security reasons")
	}

	// Resolve IP address
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If DNS lookup fails, block the URL to be safe
		return fmt.Errorf("cannot resolve hostname: %w", err)
	}

	// Check if any resolved IP is private or loopback
	for _, ip := range ips {
		if err := validateIP(ip); err != nil {
			return fmt.Errorf("hostname resolves to blocked IP %s: %w", ip, err)
		}
	}

	return nil
}

// isLocalhost checks if hostname is a localhost variation
func isLocalhost(hostname string) bool {
	localhostNames := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
		"::",
	}

	hostname = strings.ToLower(hostname)
	for _, local := range localhostNames {
		if hostname == local {
			return true
		}
	}

	return false
}

// validateIP checks if an IP address is safe (not private/loopback)
func validateIP(ip net.IP) error {
	// Check for loopback
	if ip.IsLoopback() {
		return errors.New("loopback addresses are not allowed")
	}

	// Check for private IP ranges
	if ip.IsPrivate() {
		return errors.New("private IP addresses are not allowed")
	}

	// Check for link-local addresses (169.254.x.x for IPv4, fe80::/10 for IPv6)
	if ip.IsLinkLocalUnicast() {
		return errors.New("link-local addresses are not allowed")
	}

	// Check for multicast
	if ip.IsMulticast() {
		return errors.New("multicast addresses are not allowed")
	}

	return nil
}

// ValidateCrawlerConfig validates crawler configuration parameters
func (v *Validator) ValidateCrawlerConfig(userAgent string, maxDepth, concurrency int) error {
	if userAgent == "" {
		return errors.New("user agent cannot be empty")
	}

	if maxDepth < 1 || maxDepth > 10 {
		return errors.New("max depth must be between 1 and 10")
	}

	if concurrency < 1 || concurrency > 100 {
		return errors.New("concurrency must be between 1 and 100")
	}

	return nil
}

// ValidateTimeout validates timeout values (in seconds)
func (v *Validator) ValidateTimeout(timeout int) error {
	if timeout < 1 {
		return errors.New("timeout must be at least 1 second")
	}

	if timeout > 300 {
		return errors.New("timeout cannot exceed 300 seconds (5 minutes)")
	}

	return nil
}

// ValidateDatabaseConfig validates database configuration
func (v *Validator) ValidateDatabaseConfig(host string, port int, database string) error {
	if host == "" {
		return errors.New("database host cannot be empty")
	}

	if port < 1 || port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}

	if database == "" {
		return errors.New("database name cannot be empty")
	}

	return nil
}

// ValidateCacheConfig validates cache configuration
func (v *Validator) ValidateCacheConfig(addr string, db int) error {
	if addr == "" {
		return errors.New("cache address cannot be empty")
	}

	if db < 0 || db > 15 {
		return errors.New("cache database must be between 0 and 15")
	}

	return nil
}
