package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
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

func TestNewRequestSetsAuthAndUserAgent(t *testing.T) {
	t.Parallel()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, err := c.newRequest(context.Background(), http.MethodGet, "/balance", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Header.Get("X-API-Key") != "test-key" {
		t.Fatalf("expected X-API-Key header to be set")
	}
	if req.Header.Get("User-Agent") != c.Config().UserAgent {
		t.Fatalf("expected User-Agent header to be %q", c.Config().UserAgent)
	}
	if req.Header.Get("Content-Type") != "" {
		t.Fatalf("expected empty Content-Type for nil body, got %q", req.Header.Get("Content-Type"))
	}
}

func TestNewRequestSetsContentTypeWhenBodyProvided(t *testing.T) {
	t.Parallel()

	c, err := NewClient("test-key", WithUserAgent("wallbit-go-sdk/dev"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, err := c.newRequest(context.Background(), http.MethodPost, "/transactions", strings.NewReader(`{"amount":"10"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, got %q", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("User-Agent") != "wallbit-go-sdk/dev" {
		t.Fatalf("unexpected user agent %q", req.Header.Get("User-Agent"))
	}
}

func TestSendDecodesSuccessResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Fatalf("missing api key header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	c.cfg.BaseURL = baseURL

	var out struct {
		Status string `json:"status"`
	}
	if err := c.send(context.Background(), http.MethodGet, "/balance", nil, &out); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
	if out.Status != "ok" {
		t.Fatalf("expected status ok, got %q", out.Status)
	}
}

func TestSendReturnsWallbitAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "req_123")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request","message":"invalid","code":"WB_001","details":{"field":"amount"}}`))
	}))
	defer server.Close()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	c.cfg.BaseURL = baseURL

	err = c.send(context.Background(), http.MethodGet, "/transactions", nil, nil)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "WB_001" {
		t.Fatalf("expected code WB_001, got %q", sdkErr.Code)
	}
	if sdkErr.Message != "invalid" {
		t.Fatalf("expected message invalid, got %q", sdkErr.Message)
	}
}

func TestSendReturnsFallbackErrorForMalformedBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	c, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	c.cfg.BaseURL = baseURL

	err = c.send(context.Background(), http.MethodGet, "/assets", nil, nil)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", sdkErr.StatusCode)
	}
	if sdkErr.Message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected fallback message %q, got %q", http.StatusText(http.StatusInternalServerError), sdkErr.Message)
	}
	if sdkErr.RawBody != "not-json" {
		t.Fatalf("expected raw body to be preserved")
	}
}
