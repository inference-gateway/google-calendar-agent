package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "../../bin/test-binary", "main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		if err := os.Remove("../../bin/test-binary"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	cmd = exec.Command("../../bin/test-binary", "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run version command: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "google-calendar-agent") {
		t.Errorf("Version output should contain 'google-calendar-agent', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Version:") {
		t.Errorf("Version output should contain 'Version:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Commit:") {
		t.Errorf("Version output should contain 'Commit:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Build Date:") {
		t.Errorf("Version output should contain 'Build Date:', got: %s", outputStr)
	}
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "../../bin/test-binary", "main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		if err := os.Remove("../../bin/test-binary"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	cmd = exec.Command("../../bin/test-binary", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("Help output should contain 'Usage:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "-version") {
		t.Errorf("Help output should contain '-version' flag, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "-demo") {
		t.Errorf("Help output should contain '-demo' flag, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "-gin-mode") {
		t.Errorf("Help output should contain '-gin-mode' flag, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "LOG_LEVEL") {
		t.Errorf("Help output should contain 'LOG_LEVEL' environment variable, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "GIN_MODE") {
		t.Errorf("Help output should contain 'GIN_MODE' environment variable, got: %s", outputStr)
	}
}

func TestGinModeConfiguration(t *testing.T) {
	testCases := []struct {
		name          string
		envValue      string
		flagValue     string
		expectedMode  string
		shouldContain string
	}{
		{
			name:          "default mode when no env or flag",
			envValue:      "",
			flagValue:     "",
			expectedMode:  "debug",
			shouldContain: `"mode":"debug"`,
		},
		{
			name:          "release mode from environment variable",
			envValue:      "release",
			flagValue:     "",
			expectedMode:  "release",
			shouldContain: `"mode":"release"`,
		},
		{
			name:          "test mode from environment variable",
			envValue:      "test",
			flagValue:     "",
			expectedMode:  "test",
			shouldContain: `"mode":"test"`,
		},
		{
			name:          "flag overrides environment variable",
			envValue:      "debug",
			flagValue:     "release",
			expectedMode:  "release",
			shouldContain: `"mode":"release"`,
		},
		{
			name:          "invalid mode falls back to debug",
			envValue:      "",
			flagValue:     "invalid",
			expectedMode:  "debug",
			shouldContain: `"invalidMode":"invalid"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "build", "-o", "../../bin/test-gin-mode-binary", "main.go")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build binary: %v", err)
			}
			defer func() {
				if err := os.Remove("../../bin/test-gin-mode-binary"); err != nil {
					t.Logf("Failed to remove test binary: %v", err)
				}
			}()

			args := []string{"--demo"}
			if tc.flagValue != "" {
				args = append(args, "--gin-mode="+tc.flagValue)
			}

			cmd = exec.Command("../../bin/test-gin-mode-binary", args...)
			if tc.envValue != "" {
				cmd.Env = append(os.Environ(), "GIN_MODE="+tc.envValue)
			}

			output, err := cmd.Output()
			if err != nil {
				// This is expected since the server will try to run indefinitely
				// We check the output for expected log messages
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tc.shouldContain) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.shouldContain, outputStr)
			}
		})
	}
}
