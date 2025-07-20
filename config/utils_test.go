package config

import (
	"context"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
