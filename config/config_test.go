package config

import (
	"context"
	"testing"
	"time"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load_DefaultValues(t *testing.T) {
	ctx := context.Background()

	// Test with minimal environment (demo mode)
	lookuper := envconfig.MapLookuper(map[string]string{
		"APP_DEMO_MODE": "true",
	})

	cfg, err := LoadWithLookuper(ctx, lookuper)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, "primary", cfg.Google.CalendarID)
	assert.Equal(t, false, cfg.Google.ReadOnly)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, false, cfg.Server.EnableTLS)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, true, cfg.Logging.EnableCaller)
	assert.Equal(t, true, cfg.Logging.EnableStacktrace)
	assert.Equal(t, "dev", cfg.App.Environment)
	assert.Equal(t, false, cfg.App.Debug)
	assert.Equal(t, true, cfg.App.DemoMode)
	assert.Equal(t, int64(1048576), cfg.App.MaxRequestSize)
	assert.Equal(t, time.Second*30, cfg.App.RequestTimeout)
	assert.Equal(t, time.Second*10, cfg.Server.ReadTimeout)
	assert.Equal(t, time.Second*10, cfg.Server.WriteTimeout)
	assert.Equal(t, time.Second*60, cfg.Server.IdleTimeout)
}

func TestConfig_Load_CustomValues(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		envVars  map[string]string
		expected func(*testing.T, *Config)
	}{
		{
			name: "google_configuration",
			envVars: map[string]string{
				"GOOGLE_CALENDAR_ID":        "test@example.com",
				"GOOGLE_CALENDAR_SA_JSON":   `{"type":"service_account"}`,
				"GOOGLE_CALENDAR_READ_ONLY": "true",
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test@example.com", cfg.Google.CalendarID)
				assert.Equal(t, `{"type":"service_account"}`, cfg.Google.ServiceAccountJSON)
				assert.Equal(t, true, cfg.Google.ReadOnly)
			},
		},
		{
			name: "server_configuration",
			envVars: map[string]string{
				"SERVER_PORT":         "9090",
				"SERVER_HOST":         "127.0.0.1",
				"SERVER_GIN_MODE":     "debug",
				"SERVER_ENABLE_TLS":   "true",
				"TLS_CERT_PATH":       "/cert.pem",
				"TLS_KEY_PATH":        "/key.pem",
				"SERVER_READ_TIMEOUT": "30s",
				"APP_DEMO_MODE":       "true", // To pass validation
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "9090", cfg.Server.Port)
				assert.Equal(t, "127.0.0.1", cfg.Server.Host)
				assert.Equal(t, "debug", cfg.Server.Mode)
				assert.Equal(t, true, cfg.Server.EnableTLS)
				assert.Equal(t, time.Second*30, cfg.Server.ReadTimeout)
			},
		},
		{
			name: "logging_configuration",
			envVars: map[string]string{
				"LOG_LEVEL":             "debug",
				"LOG_FORMAT":            "console",
				"LOG_OUTPUT":            "stderr",
				"LOG_ENABLE_CALLER":     "false",
				"LOG_ENABLE_STACKTRACE": "false",
				"APP_DEMO_MODE":         "true", // To pass validation
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "debug", cfg.Logging.Level)
				assert.Equal(t, "console", cfg.Logging.Format)
				assert.Equal(t, "stderr", cfg.Logging.Output)
				assert.Equal(t, false, cfg.Logging.EnableCaller)
				assert.Equal(t, false, cfg.Logging.EnableStacktrace)
			},
		},
		{
			name: "tls_configuration",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/path/to/cert.pem",
				"TLS_KEY_PATH":      "/path/to/key.pem",
				"TLS_MIN_VERSION":   "1.3",
				"TLS_CIPHER_SUITES": "TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256",
				"APP_DEMO_MODE":     "true", // To pass validation
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/path/to/cert.pem", cfg.TLS.CertPath)
				assert.Equal(t, "/path/to/key.pem", cfg.TLS.KeyPath)
				assert.Equal(t, "1.3", cfg.TLS.MinVersion)
				assert.Equal(t, "TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256", cfg.TLS.CipherSuites)
			},
		},
		{
			name: "app_configuration",
			envVars: map[string]string{
				"APP_ENVIRONMENT":         "production",
				"APP_DEBUG":               "true",
				"APP_DEMO_MODE":           "false",
				"APP_MAX_REQUEST_SIZE":    "2097152",
				"APP_REQUEST_TIMEOUT":     "60s",
				"GOOGLE_CALENDAR_SA_JSON": `{"type":"service_account"}`, // To pass validation
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "production", cfg.App.Environment)
				assert.Equal(t, true, cfg.App.Debug)
				assert.Equal(t, false, cfg.App.DemoMode)
				assert.Equal(t, int64(2097152), cfg.App.MaxRequestSize)
				assert.Equal(t, time.Minute, cfg.App.RequestTimeout)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)
			require.NoError(t, err)
			require.NotNil(t, cfg)
			tc.expected(t, cfg)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_demo_mode",
			envVars: map[string]string{
				"APP_DEMO_MODE": "true",
			},
			expectError: false,
		},
		{
			name: "valid_with_service_account",
			envVars: map[string]string{
				"GOOGLE_CALENDAR_SA_JSON": `{"type":"service_account"}`,
			},
			expectError: false,
		},
		{
			name: "valid_with_credentials_path",
			envVars: map[string]string{
				"GOOGLE_APPLICATION_CREDENTIALS": "/path/to/credentials.json",
			},
			expectError: false,
		},
		{
			name: "missing_google_credentials",
			envVars: map[string]string{
				"APP_DEMO_MODE": "false",
			},
			expectError: true,
			errorMsg:    "either GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS must be provided",
		},
		{
			name: "tls_enabled_missing_cert",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_KEY_PATH":      "/path/to/key.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectError: true,
			errorMsg:    "TLS_CERT_PATH is required when TLS is enabled",
		},
		{
			name: "tls_enabled_missing_key",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/path/to/cert.pem",
				"APP_DEMO_MODE":     "true",
			},
			expectError: true,
			errorMsg:    "TLS_KEY_PATH is required when TLS is enabled",
		},
		{
			name: "invalid_log_level",
			envVars: map[string]string{
				"LOG_LEVEL":     "invalid",
				"APP_DEMO_MODE": "true",
			},
			expectError: true,
			errorMsg:    "invalid log level 'invalid'",
		},
		{
			name: "invalid_server_mode",
			envVars: map[string]string{
				"SERVER_GIN_MODE": "invalid",
				"APP_DEMO_MODE":   "true",
			},
			expectError: true,
			errorMsg:    "invalid server mode 'invalid'",
		},
		{
			name: "invalid_tls_version",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"TLS_CERT_PATH":     "/path/to/cert.pem",
				"TLS_KEY_PATH":      "/path/to/key.pem",
				"TLS_MIN_VERSION":   "1.1",
				"APP_DEMO_MODE":     "true",
			},
			expectError: true,
			errorMsg:    "invalid TLS version '1.1'",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lookuper := envconfig.MapLookuper(tc.envVars)
			cfg, err := LoadWithLookuper(ctx, lookuper)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			}
		})
	}
}

func TestConfig_HelperMethods(t *testing.T) {
	testCases := []struct {
		name     string
		envVars  map[string]string
		testFunc func(*testing.T, *Config)
	}{
		{
			name: "get_server_address",
			envVars: map[string]string{
				"SERVER_HOST":   "127.0.0.1",
				"SERVER_PORT":   "9090",
				"APP_DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "127.0.0.1:9090", cfg.GetServerAddress())
			},
		},
		{
			name: "is_production",
			envVars: map[string]string{
				"APP_ENVIRONMENT": "production",
				"APP_DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsProduction())
				assert.False(t, cfg.IsDevelopment())
			},
		},
		{
			name: "is_development",
			envVars: map[string]string{
				"APP_ENVIRONMENT": "development",
				"APP_DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDevelopment())
				assert.False(t, cfg.IsProduction())
			},
		},
		{
			name: "is_debug_enabled_explicit",
			envVars: map[string]string{
				"APP_DEBUG":     "true",
				"APP_DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "is_debug_enabled_log_level",
			envVars: map[string]string{
				"LOG_LEVEL":     "debug",
				"APP_DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "is_debug_enabled_development",
			envVars: map[string]string{
				"APP_ENVIRONMENT": "development",
				"APP_DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "should_use_mock_service_demo",
			envVars: map[string]string{
				"APP_DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.ShouldUseMockService())
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
			tc.testFunc(t, cfg)
		})
	}
}

func TestConfig_Load_RealEnvironment(t *testing.T) {
	// This test uses the real environment and should work with minimal setup
	ctx := context.Background()

	// Set a minimal environment for the test
	t.Setenv("APP_DEMO_MODE", "true")

	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should work with demo mode
	assert.True(t, cfg.ShouldUseMockService())
}
