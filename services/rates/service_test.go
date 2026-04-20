package rates_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/services/rates"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/rates" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("source_currency"); got != "ARS" {
			t.Fatalf("expected source_currency=ARS, got %q", got)
		}
		if got := r.URL.Query().Get("dest_currency"); got != "USD" {
			t.Fatalf("expected dest_currency=USD, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"source_currency":"ARS","dest_currency":"USD","pair":"ARSUSD","rate":1481.02,"updated_at":"2026-02-25T01:50:04+00:00"}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Rates.Get(context.Background(), rates.GetRequest{
		SourceCurrency: "ARS",
		DestCurrency:   "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Pair != "ARSUSD" {
		t.Fatalf("unexpected pair %q", out.Payload.Data.Pair)
	}
	if out.Payload.Data.Rate != 1481.02 {
		t.Fatalf("unexpected rate %v", out.Payload.Data.Rate)
	}
	if out.Payload.Data.UpdatedAt == nil {
		t.Fatal("expected UpdatedAt to be non-nil")
	}
	want := time.Date(2026, 2, 25, 1, 50, 4, 0, time.FixedZone("UTC", 0))
	if !out.Payload.Data.UpdatedAt.Equal(want) {
		t.Fatalf("unexpected UpdatedAt %v", out.Payload.Data.UpdatedAt)
	}
}

func TestServiceGetIdentityPair(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"source_currency":"USD","dest_currency":"USD","pair":"USDUSD","rate":1.0,"updated_at":null}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Rates.Get(context.Background(), rates.GetRequest{
		SourceCurrency: "USD",
		DestCurrency:   "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Rate != 1.0 {
		t.Fatalf("expected rate 1.0, got %v", out.Payload.Data.Rate)
	}
	if out.Payload.Data.UpdatedAt != nil {
		t.Fatalf("expected UpdatedAt nil for identity pair, got %v", *out.Payload.Data.UpdatedAt)
	}
}

func TestServiceGetRejectsEmptyCurrencies(t *testing.T) {
	t.Parallel()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL("http://localhost"), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []rates.GetRequest{
		{SourceCurrency: "", DestCurrency: "USD"},
		{SourceCurrency: "ARS", DestCurrency: ""},
		{SourceCurrency: " ", DestCurrency: "USD"},
	}
	for _, req := range cases {
		if _, err := c.Rates.Get(context.Background(), req); !errors.Is(err, rates.ErrEmptyCurrency) {
			t.Fatalf("expected ErrEmptyCurrency for %+v, got %v", req, err)
		}
	}
}

func TestServiceGetReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Exchange rate not found for this currency pair."}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Rates.Get(context.Background(), rates.GetRequest{SourceCurrency: "XXX", DestCurrency: "USD"})
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if !wallbit.IsNotFound(apiErr) {
		t.Fatalf("expected IsNotFound, got %v", apiErr)
	}
}
