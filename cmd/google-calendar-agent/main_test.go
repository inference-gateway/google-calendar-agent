package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "test-binary", "main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		if err := os.Remove("test-binary"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	cmd = exec.Command("./test-binary", "--version")
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
	cmd := exec.Command("go", "build", "-o", "test-binary", "main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		if err := os.Remove("test-binary"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	cmd = exec.Command("./test-binary", "--help")
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
}
