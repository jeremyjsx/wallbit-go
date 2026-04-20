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
	if out.Payload.Data.CurrentPage != 2 {
		t.Fatalf("expected current_page=2, got %d", out.Payload.Data.CurrentPage)
	}
	if len(out.Payload.Data.Data) != 2 {
		t.Fatalf("expected two transactions, got %d", len(out.Payload.Data.Data))
	}
	if out.Payload.Data.Data[0].UUID != "abc" {
		t.Fatalf("unexpected uuid %q", out.Payload.Data.Data[0].UUID)
	}
	if out.Payload.Data.Data[0].ExternalAddress == nil || *out.Payload.Data.Data[0].ExternalAddress != "Juan Perez" {
		t.Fatalf("unexpected external_address in first transaction: %v", out.Payload.Data.Data[0].ExternalAddress)
	}
	if out.Payload.Data.Data[1].UUID != "def" {
		t.Fatalf("unexpected uuid %q", out.Payload.Data.Data[1].UUID)
	}
	if out.Payload.Data.Data[1].ExternalAddress != nil {
		t.Fatalf("expected nil external_address in second transaction, got %v", out.Payload.Data.Data[1].ExternalAddress)
	}
	wantFirst := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	wantSecond := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !out.Payload.Data.Data[0].CreatedAt.Equal(wantFirst) {
		t.Fatalf("Data[0].CreatedAt: got %s, want %s", out.Payload.Data.Data[0].CreatedAt, wantFirst)
	}
	if !out.Payload.Data.Data[1].CreatedAt.Equal(wantSecond) {
		t.Fatalf("Data[1].CreatedAt: got %s, want %s", out.Payload.Data.Data[1].CreatedAt, wantSecond)
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
	if out.Payload.Data.Count != 0 {
		t.Fatalf("expected count=0, got %d", out.Payload.Data.Count)
	}
}

func TestServiceListAllWalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := map[string]string{
		"1": `{"data":{"data":[{"uuid":"a","type":"TRADE","status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","source_amount":1,"dest_amount":1,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}},{"uuid":"b","type":"TRADE","status":"COMPLETED","created_at":"2024-01-02T00:00:00Z","source_amount":2,"dest_amount":2,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":3,"current_page":1,"count":5}}`,
		"2": `{"data":{"data":[{"uuid":"c","type":"TRADE","status":"COMPLETED","created_at":"2024-01-03T00:00:00Z","source_amount":3,"dest_amount":3,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}},{"uuid":"d","type":"TRADE","status":"COMPLETED","created_at":"2024-01-04T00:00:00Z","source_amount":4,"dest_amount":4,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":3,"current_page":2,"count":5}}`,
		"3": `{"data":{"data":[{"uuid":"e","type":"TRADE","status":"COMPLETED","created_at":"2024-01-05T00:00:00Z","source_amount":5,"dest_amount":5,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":3,"current_page":3,"count":5}}`,
	}
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		page := r.URL.Query().Get("page")
		body, ok := pages[page]
		if !ok {
			t.Fatalf("unexpected page requested: %q", page)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	limit := 2
	var got []string
	for tx, err := range c.Transactions.ListAll(context.Background(), &transactions.ListRequest{Limit: &limit}) {
		if err != nil {
			t.Fatalf("unexpected iteration error: %v", err)
		}
		got = append(got, tx.UUID)
	}
	want := []string{"a", "b", "c", "d", "e"}
	if len(got) != len(want) {
		t.Fatalf("expected %d transactions, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("item %d: got %q, want %q", i, got[i], want[i])
		}
	}
	if hits != 3 {
		t.Fatalf("expected 3 HTTP calls, got %d", hits)
	}
}

func TestServiceListAllStopsOnBreak(t *testing.T) {
	t.Parallel()

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":[{"uuid":"a","type":"TRADE","status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","source_amount":1,"dest_amount":1,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":10,"current_page":1,"count":10}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, err := range c.Transactions.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatalf("unexpected iteration error: %v", err)
		}
		break
	}
	if hits != 1 {
		t.Fatalf("expected iteration to stop after first page, got %d HTTP calls", hits)
	}
}

func TestServiceListAllPropagatesAPIError(t *testing.T) {
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

	var sawErr error
	for _, iterErr := range c.Transactions.ListAll(context.Background(), nil) {
		if iterErr != nil {
			sawErr = iterErr
			break
		}
		t.Fatal("expected error on first yield")
	}
	var apiErr *wallbit.Error
	if !errors.As(sawErr, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", sawErr)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}

func TestServiceListAllDoesNotMutateCallerRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":[{"uuid":"a","type":"TRADE","status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","source_amount":1,"dest_amount":1,"source_currency":{"code":"USD","alias":"USD"},"dest_currency":{"code":"USD","alias":"USD"}}],"pages":1,"current_page":1,"count":1}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origPage := 1
	req := &transactions.ListRequest{Page: &origPage, Currency: "USD"}
	for range c.Transactions.ListAll(context.Background(), req) {
	}
	if req.Page == nil || *req.Page != 1 {
		t.Fatalf("caller Page mutated: got %v", req.Page)
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
