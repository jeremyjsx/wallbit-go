package client

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	t.Parallel()

	_, err := NewClient("   ")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
	if err != ErrMissingAPIKey {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestNewClientUsesDefaults(t *testing.T) {
	t.Parallel()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := c.Config()
	if cfg.BaseURL.String() != defaultBaseURL {
		t.Fatalf("expected base url %q, got %q", defaultBaseURL, cfg.BaseURL.String())
	}
	if cfg.HTTPClient == nil {
		t.Fatal("expected default http client")
	}
	if cfg.HTTPClient.Timeout != 30*time.Second {
		t.Fatalf("expected timeout 30s, got %s", cfg.HTTPClient.Timeout)
	}
}

func TestNewClientAppliesOptions(t *testing.T) {
	t.Parallel()

	customHTTPClient := &http.Client{Timeout: 5 * time.Second}
	c, err := NewClient(
		"test-key",
		WithBaseURL("https://sandbox.wallbit.io"),
		WithHTTPClient(customHTTPClient),
		WithTimeout(2*time.Second),
		WithUserAgent("wallbit-go-sdk/dev"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := c.Config()
	if cfg.BaseURL.String() != "https://sandbox.wallbit.io" {
		t.Fatalf("unexpected base url %q", cfg.BaseURL.String())
	}
	if cfg.HTTPClient != customHTTPClient {
		t.Fatal("expected configured http client to be used")
	}
	if cfg.HTTPClient.Timeout != 2*time.Second {
		t.Fatalf("expected timeout 2s, got %s", cfg.HTTPClient.Timeout)
	}
	if cfg.UserAgent != "wallbit-go-sdk/dev" {
		t.Fatalf("unexpected user agent %q", cfg.UserAgent)
	}
}
