package wallets_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/wallets"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/wallets" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("currency") != "USDT" {
			t.Fatalf("expected currency=USDT, got %q", r.URL.Query().Get("currency"))
		}
		if r.URL.Query().Get("network") != "ethereum" {
			t.Fatalf("expected network=ethereum, got %q", r.URL.Query().Get("network"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"address":"0xabc","network":"ethereum","currency_code":"USDT"}]}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Wallets.Get(context.Background(), &wallets.GetRequest{
		Currency: "USDT",
		Network:  "ethereum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 1 {
		t.Fatalf("expected one wallet, got %d", len(out.Data))
	}
	if out.Data[0].CurrencyCode != "USDT" {
		t.Fatalf("unexpected currency_code %q", out.Data[0].CurrencyCode)
	}
}

func TestServiceGetWithoutFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Wallets.Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 0 {
		t.Fatalf("expected no wallets, got %d", len(out.Data))
	}
}

func TestServiceGetReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden","code":"INSUFFICIENT_PERMISSIONS"}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Wallets.Get(context.Background(), nil)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
