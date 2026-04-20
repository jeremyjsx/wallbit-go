package balance_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceGetChecking(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/balance/checking" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"currency":"USD","balance":100.5}]}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
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
	if out.Data[0].Currency != "USD" {
		t.Fatalf("expected currency USD, got %q", out.Data[0].Currency)
	}
	if out.Data[0].Balance != 100.5 {
		t.Fatalf("expected balance 100.5, got %v", out.Data[0].Balance)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Balance.GetStocks(context.Background())
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
