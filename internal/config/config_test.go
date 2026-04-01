package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	ResetForTest()
	// Clear any env vars that may be set in the shell
	os.Unsetenv("APPSCAN_BASE_URL")
	os.Unsetenv("APPSCAN_KEY_ID")
	os.Unsetenv("APPSCAN_KEY_SECRET")
	os.Unsetenv("APPSCAN_API_KEY")
	defer os.Unsetenv("APPSCAN_BASE_URL")
	defer os.Unsetenv("APPSCAN_KEY_ID")
	defer os.Unsetenv("APPSCAN_KEY_SECRET")
	defer os.Unsetenv("APPSCAN_API_KEY")
	cfg := Load()

	if cfg.BaseURL != "" {
		t.Errorf("expected empty base URL, got %s", cfg.BaseURL)
	}
	if cfg.KeyID != "" {
		t.Errorf("expected empty key ID, got %s", cfg.KeyID)
	}
	if cfg.KeySecret != "" {
		t.Errorf("expected empty key secret, got %s", cfg.KeySecret)
	}
	if cfg.TimeoutSeconds != 30 {
		t.Errorf("expected default timeout 30, got %d", cfg.TimeoutSeconds)
	}
	if cfg.ReadOnly() {
		t.Error("expected readonly false by default")
	}
}

func TestLoadFromEnvVars(t *testing.T) {
	ResetForTest()
	os.Setenv("APPSCAN_BASE_URL", "https://cloud.appscan.com/api/v4")
	os.Setenv("APPSCAN_KEY_ID", "test-key-id")
	os.Setenv("APPSCAN_KEY_SECRET", "test-key-secret")
	os.Setenv("APPSCAN_TIMEOUT_SECONDS", "60")
	defer os.Unsetenv("APPSCAN_BASE_URL")
	defer os.Unsetenv("APPSCAN_KEY_ID")
	defer os.Unsetenv("APPSCAN_KEY_SECRET")
	defer os.Unsetenv("APPSCAN_TIMEOUT_SECONDS")

	cfg := Load()

	if cfg.BaseURL != "https://cloud.appscan.com/api/v4" {
		t.Errorf("expected base URL https://cloud.appscan.com/api/v4, got %s", cfg.BaseURL)
	}
	if cfg.KeyID != "test-key-id" {
		t.Errorf("expected key ID test-key-id, got %s", cfg.KeyID)
	}
	if cfg.KeySecret != "test-key-secret" {
		t.Errorf("expected key secret test-key-secret, got %s", cfg.KeySecret)
	}
	if cfg.TimeoutSeconds != 60 {
		t.Errorf("expected timeout 60, got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadFromCombinedAPIKey(t *testing.T) {
	ResetForTest()
	os.Setenv("APPSCAN_BASE_URL", "https://cloud.appscan.com/api/v4")
	os.Setenv("APPSCAN_API_KEY", "key-id-abc:key-secret-xyz")
	defer os.Unsetenv("APPSCAN_BASE_URL")
	defer os.Unsetenv("APPSCAN_API_KEY")

	cfg := Load()

	if cfg.KeyID != "key-id-abc" {
		t.Errorf("expected key ID key-id-abc, got %s", cfg.KeyID)
	}
	if cfg.KeySecret != "key-secret-xyz" {
		t.Errorf("expected key secret key-secret-xyz, got %s", cfg.KeySecret)
	}
}

func TestCombinedAPIKeyTakesPrecedence(t *testing.T) {
	ResetForTest()
	os.Setenv("APPSCAN_BASE_URL", "https://cloud.appscan.com/api/v4")
	os.Setenv("APPSCAN_KEY_ID", "should-be-ignored")
	os.Setenv("APPSCAN_KEY_SECRET", "should-be-ignored")
	os.Setenv("APPSCAN_API_KEY", "combined-id:combined-secret")
	defer os.Unsetenv("APPSCAN_BASE_URL")
	defer os.Unsetenv("APPSCAN_KEY_ID")
	defer os.Unsetenv("APPSCAN_KEY_SECRET")
	defer os.Unsetenv("APPSCAN_API_KEY")

	cfg := Load()

	if cfg.KeyID != "combined-id" {
		t.Errorf("expected key ID combined-id, got %s", cfg.KeyID)
	}
	if cfg.KeySecret != "combined-secret" {
		t.Errorf("expected key secret combined-secret, got %s", cfg.KeySecret)
	}
}
