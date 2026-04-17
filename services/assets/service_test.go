package assets_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/assets"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %q", r.Method)
		}
		wantPath := "/api/public/v1/assets/" + url.PathEscape("BRK.B")
		if r.URL.Path != wantPath {
			t.Fatalf("unexpected path %q, want %q", r.URL.Path, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"symbol":"BRK.B","name":"Berkshire Hathaway Inc.","price":412.3,"logo_url":"https://example.com/BRK.B.svg"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Assets.Get(context.Background(), "BRK.B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.Symbol != "BRK.B" {
		t.Fatalf("unexpected symbol %q", out.Data.Symbol)
	}
	if out.Data.Name != "Berkshire Hathaway Inc." {
		t.Fatalf("unexpected name %q", out.Data.Name)
	}
}

func TestServiceGetNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Asset not found","code":"NOT_FOUND"}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Assets.Get(context.Background(), "NOPE")
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status %d", sdkErr.StatusCode)
	}
}

func TestServiceList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %q", r.Method)
		}
		if r.URL.Path != "/api/public/v1/assets" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("category") != "TECHNOLOGY" {
			t.Fatalf("expected category=TECHNOLOGY, got %q", r.URL.Query().Get("category"))
		}
		if r.URL.Query().Get("search") != "Apple" {
			t.Fatalf("expected search=Apple, got %q", r.URL.Query().Get("search"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "20" {
			t.Fatalf("expected limit=20, got %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"symbol":"AAPL","name":"Apple Inc.","price":175.5,"asset_type":"Stock","exchange":"NASDAQ","sector":"Technology","market_cap_m":"2750000","description":"Apple description","description_es":"Descripcion Apple","country":"United States","ceo":"Tim Cook","employees":"164000","logo_url":"https://static.atomicvest.com/AAPL.svg","dividend":{"amount":0.24,"yield":0.52,"ex_date":"2024-02-09","payment_date":"2024-02-15"}}],"pages":15,"current_page":2,"count":150}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	page := 2
	limit := 20
	out, err := c.Assets.List(context.Background(), &assets.ListRequest{
		Category: "TECHNOLOGY",
		Search:   "Apple",
		Page:     &page,
		Limit:    &limit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CurrentPage != 2 {
		t.Fatalf("expected current_page=2, got %d", out.CurrentPage)
	}
	if len(out.Data) != 1 {
		t.Fatalf("expected one asset, got %d", len(out.Data))
	}
	if out.Data[0].Symbol != "AAPL" {
		t.Fatalf("unexpected symbol %q", out.Data[0].Symbol)
	}
	if out.Data[0].Dividend == nil {
		t.Fatal("expected dividend data")
	}
	if out.Data[0].Dividend.Amount == nil || *out.Data[0].Dividend.Amount != 0.24 {
		t.Fatalf("unexpected dividend amount: %v", out.Data[0].Dividend.Amount)
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
		_, _ = w.Write([]byte(`{"data":[],"pages":0,"current_page":1,"count":0}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Assets.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Count != 0 {
		t.Fatalf("expected count=0, got %d", out.Count)
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

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Assets.List(context.Background(), nil)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
