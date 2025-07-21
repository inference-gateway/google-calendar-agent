package config

import (
	"context"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load_DefaultValues(t *testing.T) {
	ctx := context.Background()

	lookuper := envconfig.MapLookuper(map[string]string{
		"DEMO_MODE":   "true",
		"ENVIRONMENT": "prod",
	})

	cfg, err := LoadWithLookuper(ctx, lookuper)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "primary", cfg.Google.CalendarID)
	assert.Equal(t, false, cfg.Google.ReadOnly)
	assert.Equal(t, "8080", cfg.A2A.ServerConfig.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, true, cfg.Logging.EnableCaller)
	assert.Equal(t, true, cfg.Logging.EnableStacktrace)
	assert.Equal(t, "prod", cfg.Environment)
	assert.Equal(t, false, cfg.IsDebugEnabled())
	assert.Equal(t, true, cfg.DemoMode)
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
				"A2A_SERVER_PORT": "9090",
				"DEMO_MODE":       "true",
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "9090", cfg.A2A.ServerConfig.Port)
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
				"DEMO_MODE":             "true",
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
			name: "app_configuration",
			envVars: map[string]string{
				"ENVIRONMENT":                    "production",
				"LOG_LEVEL":                      "debug",
				"DEMO_MODE":                      "false",
				"GOOGLE_APPLICATION_CREDENTIALS": `{"type":"service_account"}`,
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "production", cfg.Environment)
				assert.Equal(t, true, cfg.IsDebugEnabled())
				assert.Equal(t, false, cfg.DemoMode)
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
				"DEMO_MODE": "true",
			},
			expectError: false,
		},
		{
			name: "valid_with_service_account",
			envVars: map[string]string{
				"GOOGLE_APPLICATION_CREDENTIALS": `{"type":"service_account"}`,
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
				"DEMO_MODE": "false",
			},
			expectError: true,
			errorMsg:    "either GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS must be provided when not in demo mode",
		},
		{
			name: "invalid_log_level",
			envVars: map[string]string{
				"LOG_LEVEL": "invalid",
				"DEMO_MODE": "true",
			},
			expectError: true,
			errorMsg:    "invalid log level 'invalid'",
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
				"A2A_SERVER_PORT": "9090",
				"DEMO_MODE":       "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":9090", cfg.GetServerAddress())
			},
		},
		{
			name: "is_production",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsProduction())
				assert.False(t, cfg.IsDevelopment())
			},
		},
		{
			name: "is_development",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDevelopment())
				assert.False(t, cfg.IsProduction())
			},
		},
		{
			name: "is_debug_enabled_explicit",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
				"DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "is_debug_enabled_log_level",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
				"DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "is_debug_enabled_development",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"DEMO_MODE":   "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.IsDebugEnabled())
			},
		},
		{
			name: "should_use_mock_service_demo",
			envVars: map[string]string{
				"DEMO_MODE": "true",
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
	ctx := context.Background()

	t.Setenv("DEMO_MODE", "true")

	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.ShouldUseMockService())
}
