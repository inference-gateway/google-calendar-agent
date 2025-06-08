package config

import (
	"crypto/tls"
	"fmt"
	"strings"
)

// GetTLSConfig returns a TLS configuration based on the config settings
func (c *Config) GetTLSConfig() (*tls.Config, error) {
	if !c.Server.EnableTLS {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	switch c.TLS.MinVersion {
	case "1.2":
		tlsConfig.MinVersion = tls.VersionTLS12
	case "1.3":
		tlsConfig.MinVersion = tls.VersionTLS13
	default:
		return nil, fmt.Errorf("unsupported TLS version: %s", c.TLS.MinVersion)
	}

	if c.TLS.CipherSuites != "" {
		ciphers := strings.Split(c.TLS.CipherSuites, ",")
		var cipherSuites []uint16

		for _, cipher := range ciphers {
			cipher = strings.TrimSpace(cipher)
			switch cipher {
			case "TLS_AES_128_GCM_SHA256":
				cipherSuites = append(cipherSuites, tls.TLS_AES_128_GCM_SHA256)
			case "TLS_AES_256_GCM_SHA384":
				cipherSuites = append(cipherSuites, tls.TLS_AES_256_GCM_SHA384)
			case "TLS_CHACHA20_POLY1305_SHA256":
				cipherSuites = append(cipherSuites, tls.TLS_CHACHA20_POLY1305_SHA256)
			case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
			case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
			case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
			case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
			case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305)
			case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":
				cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)
			default:
				return nil, fmt.Errorf("unsupported cipher suite: %s", cipher)
			}
		}

		if len(cipherSuites) > 0 {
			tlsConfig.CipherSuites = cipherSuites
		}
	}

	return tlsConfig, nil
}

// GetGoogleCredentialsOption returns the appropriate Google API credential option
func (c *Config) GetGoogleCredentialsOption() (string, string, error) {
	if c.ShouldUseMockService() {
		return "", "", nil
	}

	if c.Google.ServiceAccountJSON != "" {
		return "json", c.Google.ServiceAccountJSON, nil
	}

	if c.Google.CredentialsPath != "" {
		return "file", c.Google.CredentialsPath, nil
	}

	return "", "", fmt.Errorf("no Google credentials configured")
}

// GetLogLevel returns the zap log level equivalent
func (c *Config) GetLogLevel() string {
	// Convert our log level to match what the application expects
	switch strings.ToLower(c.Logging.Level) {
	case "debug":
		return "debug"
	case "info":
		return "info"
	case "warn", "warning":
		return "warn"
	case "error":
		return "error"
	default:
		return "info"
	}
}

// GetPort returns the port with appropriate protocol prefix
func (c *Config) GetPort() string {
	if c.Server.EnableTLS && c.Server.Port == "8080" {
		return "8443"
	}
	return c.Server.Port
}

// GetProtocol returns the protocol scheme (http or https)
func (c *Config) GetProtocol() string {
	if c.Server.EnableTLS {
		return "https"
	}
	return "http"
}

// GetBaseURL returns the complete base URL for the server
func (c *Config) GetBaseURL() string {
	protocol := c.GetProtocol()
	port := c.GetPort()

	if c.Server.Host == "localhost" || c.Server.Host == "127.0.0.1" || c.Server.Host == "0.0.0.0" {
		return fmt.Sprintf("%s://%s:%s", protocol, c.Server.Host, port)
	}

	if (protocol == "http" && port == "80") || (protocol == "https" && port == "443") {
		return fmt.Sprintf("%s://%s", protocol, c.Server.Host)
	}

	return fmt.Sprintf("%s://%s:%s", protocol, c.Server.Host, port)
}

// ToMap converts the config to a map for debugging/logging purposes
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"google": map[string]interface{}{
			"calendar_id":     c.Google.CalendarID,
			"read_only":       c.Google.ReadOnly,
			"has_credentials": c.Google.ServiceAccountJSON != "" || c.Google.CredentialsPath != "",
		},
		"server": map[string]interface{}{
			"port":          c.Server.Port,
			"host":          c.Server.Host,
			"mode":          c.Server.Mode,
			"enable_tls":    c.Server.EnableTLS,
			"read_timeout":  c.Server.ReadTimeout.String(),
			"write_timeout": c.Server.WriteTimeout.String(),
			"idle_timeout":  c.Server.IdleTimeout.String(),
		},
		"logging": map[string]interface{}{
			"level":             c.Logging.Level,
			"format":            c.Logging.Format,
			"output":            c.Logging.Output,
			"enable_caller":     c.Logging.EnableCaller,
			"enable_stacktrace": c.Logging.EnableStacktrace,
		},
		"tls": map[string]interface{}{
			"cert_path":     c.TLS.CertPath,
			"key_path":      c.TLS.KeyPath,
			"min_version":   c.TLS.MinVersion,
			"cipher_suites": c.TLS.CipherSuites,
		},
		"app": map[string]interface{}{
			"environment":      c.App.Environment,
			"debug":            c.App.Debug,
			"demo_mode":        c.App.DemoMode,
			"max_request_size": c.App.MaxRequestSize,
			"request_timeout":  c.App.RequestTimeout.String(),
		},
	}
}
