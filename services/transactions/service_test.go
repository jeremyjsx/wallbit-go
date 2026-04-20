package transactions_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/services/transactions"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/v1/transactions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "20" {
			t.Fatalf("expected limit=20, got %q", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("currency") != "USD" {
			t.Fatalf("expected currency=USD, got %q", r.URL.Query().Get("currency"))
		}
		if r.URL.Query().Get("from_date") != "2024-01-01" {
			t.Fatalf("unexpected from_date %q", r.URL.Query().Get("from_date"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":[{"uuid":"abc","type":"TRADE","external_address":"Juan Perez","source_amount":100,"dest_amount":100,"status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}},{"uuid":"def","type":"WITHDRAWAL_LOCAL","external_address":null,"source_amount":50,"dest_amount":50,"status":"PENDING","created_at":"2024-01-02T00:00:00Z","source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":5,"current_page":2,"count":50}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	page := 2
	limit := 20
	fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	out, err := c.Transactions.List(context.Background(), &transactions.ListRequest{
		Page:     &page,
		Limit:    &limit,
		Currency: "USD",
		FromDate: &fromDate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.CurrentPage != 2 {
		t.Fatalf("expected current_page=2, got %d", out.Data.CurrentPage)
	}
	if len(out.Data.Data) != 2 {
		t.Fatalf("expected two transactions, got %d", len(out.Data.Data))
	}
	if out.Data.Data[0].UUID != "abc" {
		t.Fatalf("unexpected uuid %q", out.Data.Data[0].UUID)
	}
	if out.Data.Data[0].ExternalAddress == nil || *out.Data.Data[0].ExternalAddress != "Juan Perez" {
		t.Fatalf("unexpected external_address in first transaction: %v", out.Data.Data[0].ExternalAddress)
	}
	if out.Data.Data[1].UUID != "def" {
		t.Fatalf("unexpected uuid %q", out.Data.Data[1].UUID)
	}
	if out.Data.Data[1].ExternalAddress != nil {
		t.Fatalf("expected nil external_address in second transaction, got %v", out.Data.Data[1].ExternalAddress)
	}
}

func TestServiceListWithoutFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":[],"pages":0,"current_page":1,"count":0}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Transactions.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.Count != 0 {
		t.Fatalf("expected count=0, got %d", out.Data.Count)
	}
}

func TestServiceListReturnsAPIError(t *testing.T) {
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

	_, err = c.Transactions.List(context.Background(), nil)
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
