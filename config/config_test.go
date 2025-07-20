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

	lookuper := envconfig.MapLookuper(map[string]string{
		"APP_DEMO_MODE":   "true",
		"APP_ENVIRONMENT": "prod",
	})

	cfg, err := LoadWithLookuper(ctx, lookuper)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "primary", cfg.Google.CalendarID)
	assert.Equal(t, false, cfg.Google.ReadOnly)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, false, cfg.Server.EnableTLS)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, true, cfg.Logging.EnableCaller)
	assert.Equal(t, true, cfg.Logging.EnableStacktrace)
	assert.Equal(t, "prod", cfg.App.Environment)
	assert.Equal(t, false, cfg.IsDebugEnabled())
	assert.Equal(t, true, cfg.App.DemoMode)
	assert.Equal(t, int64(1048576), cfg.App.MaxRequestSize)
	assert.Equal(t, time.Second*30, cfg.App.RequestTimeout)
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
				"SERVER_PORT":   "9090",
				"APP_DEMO_MODE": "true",
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "9090", cfg.Server.Port)
				assert.Equal(t, false, cfg.Server.EnableTLS)
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
				"APP_DEMO_MODE":         "true",
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
				"APP_ENVIRONMENT":                "production",
				"LOG_LEVEL":                      "debug",
				"APP_DEMO_MODE":                  "false",
				"APP_MAX_REQUEST_SIZE":           "2097152",
				"APP_REQUEST_TIMEOUT":            "60s",
				"GOOGLE_APPLICATION_CREDENTIALS": `{"type":"service_account"}`,
			},
			expected: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "production", cfg.App.Environment)
				assert.Equal(t, true, cfg.IsDebugEnabled())
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
				"APP_DEMO_MODE": "false",
			},
			expectError: true,
			errorMsg:    "GOOGLE_APPLICATION_CREDENTIALS must be provided when not in demo mode",
		},
		{
			name: "tls_enabled_but_should_use_adk",
			envVars: map[string]string{
				"SERVER_ENABLE_TLS": "true",
				"APP_DEMO_MODE":     "true",
			},
			expectError: true,
			errorMsg:    "TLS configuration is handled by the A2A ADK framework",
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

func TestConfig_Validate_LLM(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "llm_disabled_no_validation",
			envVars: map[string]string{
				"APP_DEMO_MODE": "true",
				"LLM_ENABLED":   "false",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_valid_openai",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
				"LLM_TEMPERATURE": "0.7",
				"LLM_MAX_TOKENS":  "2048",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_valid_groq",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "groq",
				"LLM_MODEL":       "deepseek-r1-distill-llama-70b",
				"LLM_TEMPERATURE": "0.5",
				"LLM_MAX_TOKENS":  "4096",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_valid_anthropic",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "anthropic",
				"LLM_MODEL":       "claude-3-opus",
				"LLM_TEMPERATURE": "1.0",
				"LLM_MAX_TOKENS":  "1024",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_empty_gateway_url",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
			},
			expectError: true,
			errorMsg:    "LLM_GATEWAY_URL is required when LLM is enabled",
		},
		{
			name: "llm_enabled_empty_provider",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "",
				"LLM_MODEL":       "gpt-4o",
			},
			expectError: true,
			errorMsg:    "LLM_PROVIDER is required when LLM is enabled",
		},
		{
			name: "llm_enabled_invalid_provider",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "invalid-provider",
				"LLM_MODEL":       "gpt-4o",
			},
			expectError: true,
			errorMsg:    "invalid LLM provider 'invalid-provider', must be one of: openai, anthropic, groq, ollama, deepseek, cohere, cloudflare",
		},
		{
			name: "llm_enabled_empty_model",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "",
			},
			expectError: true,
			errorMsg:    "LLM_MODEL is required when LLM is enabled",
		},
		{
			name: "llm_enabled_invalid_temperature_low",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
				"LLM_TEMPERATURE": "-0.1",
			},
			expectError: true,
			errorMsg:    "LLM_TEMPERATURE must be between 0.0 and 2.0, got -0.100000",
		},
		{
			name: "llm_enabled_invalid_temperature_high",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
				"LLM_TEMPERATURE": "2.1",
			},
			expectError: true,
			errorMsg:    "LLM_TEMPERATURE must be between 0.0 and 2.0, got 2.100000",
		},
		{
			name: "llm_enabled_invalid_max_tokens",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
				"LLM_MAX_TOKENS":  "0",
			},
			expectError: true,
			errorMsg:    "LLM_MAX_TOKENS must be greater than 0, got 0",
		},
		{
			name: "llm_enabled_valid_all_providers",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "deepseek",
				"LLM_MODEL":       "deepseek-r1-distill-llama-70b",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_cloudflare_provider",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "cloudflare",
				"LLM_MODEL":       "@cf/meta/llama-3.1-8b-instruct",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_ollama_provider",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "ollama",
				"LLM_MODEL":       "llama3.2",
			},
			expectError: false,
		},
		{
			name: "llm_enabled_cohere_provider",
			envVars: map[string]string{
				"APP_DEMO_MODE":   "true",
				"LLM_ENABLED":     "true",
				"LLM_GATEWAY_URL": "http://localhost:8080/v1",
				"LLM_PROVIDER":    "cohere",
				"LLM_MODEL":       "command-r-plus",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			lookuper := envconfig.MapLookuper(tc.envVars)

			cfg, err := LoadWithLookuper(ctx, lookuper)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)

				if tc.envVars["LLM_ENABLED"] == "true" {
					assert.True(t, cfg.LLM.Enabled)
					assert.Equal(t, tc.envVars["LLM_GATEWAY_URL"], cfg.LLM.GatewayURL)
					assert.Equal(t, tc.envVars["LLM_PROVIDER"], cfg.LLM.Provider)
					assert.Equal(t, tc.envVars["LLM_MODEL"], cfg.LLM.Model)
				}
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
				"SERVER_PORT":   "9090",
				"APP_DEMO_MODE": "true",
			},
			testFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, ":9090", cfg.GetServerAddress())
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
				"LOG_LEVEL":     "debug",
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
	ctx := context.Background()

	t.Setenv("APP_DEMO_MODE", "true")

	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.ShouldUseMockService())
}
