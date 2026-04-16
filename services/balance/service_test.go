package balance_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
)

func TestServiceGetChecking(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/balance/checking" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"currency":"USD","available_balance":100.5}]}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Balance.GetChecking(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 1 {
		t.Fatalf("expected one balance row, got %d", len(out.Data))
	}
	if out.Data[0]["currency"] != "USD" {
		t.Fatalf("expected currency USD, got %v", out.Data[0]["currency"])
	}
}

func TestServiceGetStocks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/balance/stocks" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"symbol":"AAPL","shares":10.5}]}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Balance.GetStocks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 1 {
		t.Fatalf("expected one stock position, got %d", len(out.Data))
	}
	if out.Data[0].Symbol != "AAPL" {
		t.Fatalf("expected symbol AAPL, got %q", out.Data[0].Symbol)
	}
}

func TestServiceGetStocksReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden","code":"INSUFFICIENT_PERMISSIONS"}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Balance.GetStocks(context.Background())
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
