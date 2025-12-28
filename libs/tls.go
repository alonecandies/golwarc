package libs

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// TLSConfig holds TLS/SSL configuration
type TLSConfig struct {
	Enabled            bool
	CACert             string
	ClientCert         string
	ClientKey          string
	InsecureSkipVerify bool
}

// CreateTLSConfig creates a TLS configuration from TLSConfig
// This is a shared utility used by Redis, MySQL, PostgreSQL, etc.
func CreateTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	// Load CA certificate if provided
	if cfg.CACert != "" {
		if err := loadCACertificate(tlsConfig, cfg.CACert); err != nil {
			return nil, err
		}
	}

	// Load client certificate and key if provided
	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		if err := loadClientCertificate(tlsConfig, cfg.ClientCert, cfg.ClientKey); err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

// loadCACertificate loads and parses a CA certificate
func loadCACertificate(tlsConfig *tls.Config, caCertPath string) error {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return errors.New("failed to parse CA certificate")
	}

	tlsConfig.RootCAs = caCertPool
	return nil
}

// loadClientCertificate loads a client certificate and key pair
func loadClientCertificate(tlsConfig *tls.Config, certPath, keyPath string) error {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig.Certificates = []tls.Certificate{cert}
	return nil
}
