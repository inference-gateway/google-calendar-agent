package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	testBinaryPath string
	testCertPath   string
	testKeyPath    string
)

// TestMain sets up shared resources for all tests
func TestMain(m *testing.M) {
	var err error
	testBinaryPath, err = buildTestBinary()
	if err != nil {
		os.Exit(1)
	}
	defer os.Remove(testBinaryPath)

	testCertPath, testKeyPath, err = createTestCertificates()
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		os.Remove(testCertPath)
		os.Remove(testKeyPath)
	}()

	code := m.Run()
	os.Exit(code)
}

// buildTestBinary builds the test binary once and returns its path
func buildTestBinary() (string, error) {
	binaryPath := filepath.Join("../../bin", "test-binary")
	cmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	return binaryPath, cmd.Run()
}

// createTestCertificates creates certificates once for all tests
func createTestCertificates() (certPath, keyPath string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:              []string{"localhost"},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	certFile, err := os.CreateTemp("", "test-cert-*.crt")
	if err != nil {
		return "", "", err
	}
	certPath = certFile.Name()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certFile.Close()
		os.Remove(certPath)
		return "", "", err
	}
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.key")
	if err != nil {
		os.Remove(certPath)
		return "", "", err
	}
	keyPath = keyFile.Name()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		keyFile.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", err
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		keyFile.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", err
	}
	keyFile.Close()

	return certPath, keyPath, nil
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run version command: %v", err)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"google-calendar-agent",
		"Version:",
		"Commit:",
		"Build Date:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Version output should contain '%s', got: %s", expected, outputStr)
		}
	}
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"Usage:",
		"-version",
		"-demo",
		"-gin-mode",
		"LOG_LEVEL",
		"GIN_MODE",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output should contain '%s', got: %s", expected, outputStr)
		}
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
			expectedMode:  "release",
			shouldContain: `"mode":"release"`,
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
			args := []string{"--demo"}
			if tc.flagValue != "" {
				args = append(args, "--gin-mode="+tc.flagValue)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, testBinaryPath, args...)

			cleanEnv := []string{}
			for _, env := range os.Environ() {
				if !strings.HasPrefix(env, "ENABLE_TLS=") &&
					!strings.HasPrefix(env, "TLS_CERT_PATH=") &&
					!strings.HasPrefix(env, "TLS_KEY_PATH=") &&
					!strings.HasPrefix(env, "GIN_MODE=") {
					cleanEnv = append(cleanEnv, env)
				}
			}
			cmd.Env = cleanEnv

			if tc.envValue != "" {
				cmd.Env = append(cmd.Env, "GIN_MODE="+tc.envValue)
			}
			cmd.Env = append(cmd.Env, "TLS_CERT_PATH="+testCertPath)
			cmd.Env = append(cmd.Env, "TLS_KEY_PATH="+testKeyPath)

			output, err := cmd.Output()
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					// This is expected - the server runs indefinitely
				} else {
					t.Logf("Command execution error (might be expected): %v", err)
				}
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tc.shouldContain) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.shouldContain, outputStr)
			}
		})
	}
}
