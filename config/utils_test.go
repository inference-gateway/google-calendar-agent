package config

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetTLSConfig(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		checkFunc   func(*testing.T, *tls.Config)
	}{
		{
			name: "tls_disabled",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *tls.Config) {
				assert.Nil(t, cfg)
			},
		},
		{
			name: "tls_enabled_default",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
			},
		},
		{
			name: "tls_version_1_3",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"TLS_MIN_VERSION":   "1.3",
				"APP_DEMO_MODE":     "true",
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Equal(t, uint16(tls.VersionTLS13), cfg.MinVersion)
			},
		},
		{
			name: "cipher_suites",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"TLS_CIPHER_SUITES": "TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256",
				"APP_DEMO_MODE":     "true",
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Len(t, cfg.CipherSuites, 2)
				assert.Contains(t, cfg.CipherSuites, uint16(tls.TLS_AES_256_GCM_SHA384))
				assert.Contains(t, cfg.CipherSuites, uint16(tls.TLS_CHACHA20_POLY1305_SHA256))
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			tlsConfig, err := cfg.GetTLSConfig()
			assert.NoError(t, err)
			tc.checkFunc(t, tlsConfig)
		})
	}

	t.Run("invalid_tls_version_direct", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{EnableTLS: true},
			TLS:    TLSConfig{MinVersion: "1.1"},
		}

		_, err := cfg.GetTLSConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported TLS version: 1.1")
	})

	t.Run("invalid_cipher_suite_direct", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{EnableTLS: true},
			TLS:    TLSConfig{MinVersion: "1.2", CipherSuites: "INVALID_CIPHER"},
		}

		_, err := cfg.GetTLSConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported cipher suite: INVALID_CIPHER")
	})
}

func TestConfig_GetGoogleCredentialsOption(t *testing.T) {
	testCases := []struct {
		name          string
		envVars       map[string]string
		expectedType  string
		expectedValue string
		expectError   bool
	}{
		{
			name: "demo_mode",
			envVars: map[string]string{
				"APP_DEMO_MODE": "true",
			},
			expectedType:  "",
			expectedValue: "",
			expectError:   false,
		},
		{
			name: "service_account_json",
			envVars: map[string]string{
				"GOOGLE_CALENDAR_SA_JSON": `{"type":"service_account"}`,
			},
			expectedType:  "json",
			expectedValue: `{"type":"service_account"}`,
			expectError:   false,
		},
		{
			name: "credentials_file",
			envVars: map[string]string{
				"GOOGLE_APPLICATION_CREDENTIALS": "/path/to/credentials.json",
			},
			expectedType:  "file",
			expectedValue: "/path/to/credentials.json",
			expectError:   false,
		},
		{
			name: "no_credentials",
			envVars: map[string]string{
				"APP_DEMO_MODE": "false",
			},
			expectedType:  "",
			expectedValue: "",
			expectError:   true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			if tc.expectError && err != nil {
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			credType, credValue, err := cfg.GetGoogleCredentialsOption()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedType, credType)
				assert.Equal(t, tc.expectedValue, credValue)
			}
		})
	}
}

func TestConfig_GetLogLevel(t *testing.T) {
	testCases := []struct {
		name          string
		envVars       map[string]string
		expectedLevel string
	}{
		{
			name: "debug",
			envVars: map[string]string{
				"LOG_LEVEL":     "debug",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "debug",
		},
		{
			name: "info",
			envVars: map[string]string{
				"LOG_LEVEL":     "info",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "info",
		},
		{
			name: "warn",
			envVars: map[string]string{
				"LOG_LEVEL":     "warn",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "warn",
		},
		{
			name: "warning",
			envVars: map[string]string{
				"LOG_LEVEL":     "warning",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "warn",
		},
		{
			name: "error",
			envVars: map[string]string{
				"LOG_LEVEL":     "error",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "error",
		},
		{
			name: "invalid_defaults_to_info",
			envVars: map[string]string{
				"LOG_LEVEL":     "invalid",
				"APP_DEMO_MODE": "true",
			},
			expectedLevel: "info",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			if err != nil {
				return
			}
			require.NotNil(t, cfg)

			level := cfg.GetLogLevel()
			assert.Equal(t, tc.expectedLevel, level)
		})
	}
}

func TestConfig_GetPort(t *testing.T) {
	testCases := []struct {
		name         string
		envVars      map[string]string
		expectedPort string
	}{
		{
			name: "http_default",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectedPort: "8080",
		},
		{
			name: "https_default",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectedPort: "8443",
		},
		{
			name: "custom_port",
			envVars: map[string]string{
				"SERVER_PORT":       "9090",
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectedPort: "9090",
		},
		{
			name: "custom_port_with_tls",
			envVars: map[string]string{
				"SERVER_PORT":       "9443",
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectedPort: "9443",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			port := cfg.GetPort()
			assert.Equal(t, tc.expectedPort, port)
		})
	}
}

func TestConfig_GetBaseURL(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectedURL string
	}{
		{
			name: "localhost_http",
			envVars: map[string]string{
				"SERVER_HOST":       "localhost",
				"SERVER_PORT":       "8080",
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectedURL: "http://localhost:8080",
		},
		{
			name: "localhost_https",
			envVars: map[string]string{
				"SERVER_HOST":       "localhost",
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectedURL: "https://localhost:8443",
		},
		{
			name: "production_http_standard_port",
			envVars: map[string]string{
				"SERVER_HOST":       "example.com",
				"SERVER_PORT":       "80",
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectedURL: "http://example.com",
		},
		{
			name: "production_https_standard_port",
			envVars: map[string]string{
				"SERVER_HOST":       "example.com",
				"SERVER_PORT":       "443",
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/cert.pem",
				"TLS_KEY_PATH":      "/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectedURL: "https://example.com",
		},
		{
			name: "production_custom_port",
			envVars: map[string]string{
				"SERVER_HOST":       "example.com",
				"SERVER_PORT":       "8080",
				"SERVER_ENABLE_TLS": "false",
				"APP_DEMO_MODE":     "true",
			},
			expectedURL: "http://example.com:8080",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			baseURL := cfg.GetBaseURL()
			assert.Equal(t, tc.expectedURL, baseURL)
		})
	}
}

func TestConfig_ToMap(t *testing.T) {
	ctx := context.Background()
	envVars := map[string]string{
		"APP_DEMO_MODE":           "true",
		"GOOGLE_CALENDAR_ID":      "test@example.com",
		"GOOGLE_CALENDAR_SA_JSON": `{"type":"service_account"}`,
		"SERVER_HOST":             "localhost",
		"SERVER_PORT":             "8080",
		"LOG_LEVEL":               "debug",
	}

	lookuper := envconfig.MapLookuper(envVars)
	cfg, err := LoadWithLookuper(ctx, lookuper)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	configMap := cfg.ToMap()

	assert.Contains(t, configMap, "google")
	assert.Contains(t, configMap, "server")
	assert.Contains(t, configMap, "logging")
	assert.Contains(t, configMap, "tls")
	assert.Contains(t, configMap, "app")

	googleConfig := configMap["google"].(map[string]interface{})
	assert.Equal(t, "test@example.com", googleConfig["calendar_id"])
	assert.Equal(t, true, googleConfig["has_credentials"])

	appConfig := configMap["app"].(map[string]interface{})
	assert.Equal(t, "dev", appConfig["environment"])
	assert.Equal(t, true, appConfig["demo_mode"])
}
