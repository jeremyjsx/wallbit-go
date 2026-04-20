package assets_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jeremyjsx/wallbit-go/services/assets"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceGetRejectsEmptySymbol(t *testing.T) {
	t.Parallel()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL("http://127.0.0.1:9"), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Assets.Get(context.Background(), "")
	if !errors.Is(err, assets.ErrEmptySymbol) {
		t.Fatalf("expected ErrEmptySymbol, got %v", err)
	}
	_, err = c.Assets.Get(context.Background(), "  \t ")
	if !errors.Is(err, assets.ErrEmptySymbol) {
		t.Fatalf("expected ErrEmptySymbol for whitespace, got %v", err)
	}
}

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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Assets.Get(context.Background(), "BRK.B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Symbol != "BRK.B" {
		t.Fatalf("unexpected symbol %q", out.Payload.Data.Symbol)
	}
	if out.Payload.Data.Name != "Berkshire Hathaway Inc." {
		t.Fatalf("unexpected name %q", out.Payload.Data.Name)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Assets.Get(context.Background(), "NOPE")
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status %d", apiErr.StatusCode)
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
		_, _ = w.Write([]byte(`{"data":[{"symbol":"AAPL","name":"Apple Inc.","price":175.5,"asset_type":"Stock","exchange":"NASDAQ","sector":"Technology","market_cap_m":"2750000","description":"Apple description","description_es":"Descripción Apple","country":"United States","ceo":"Tim Cook","employees":"164000","logo_url":"https://static.atomicvest.com/AAPL.svg","dividend":{"amount":0.24,"yield":0.52,"ex_date":"2024-02-09","payment_date":"2024-02-15"}}],"pages":15,"current_page":2,"count":150}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
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
	if out.Payload.CurrentPage != 2 {
		t.Fatalf("expected current_page=2, got %d", out.Payload.CurrentPage)
	}
	if len(out.Payload.Data) != 1 {
		t.Fatalf("expected one asset, got %d", len(out.Payload.Data))
	}
	if out.Payload.Data[0].Symbol != "AAPL" {
		t.Fatalf("unexpected symbol %q", out.Payload.Data[0].Symbol)
	}
	if out.Payload.Data[0].Dividend == nil {
		t.Fatal("expected dividend data")
	}
	if out.Payload.Data[0].Dividend.Amount == nil || *out.Payload.Data[0].Dividend.Amount != 0.24 {
		t.Fatalf("unexpected dividend amount: %v", out.Payload.Data[0].Dividend.Amount)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Assets.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Count != 0 {
		t.Fatalf("expected count=0, got %d", out.Payload.Count)
	}
}

func TestServiceListAllWalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := map[string]string{
		"1": `{"data":[{"symbol":"AAPL","name":"Apple Inc.","price":175.5,"logo_url":"u"},{"symbol":"MSFT","name":"Microsoft","price":400,"logo_url":"u"}],"pages":3,"current_page":1,"count":5}`,
		"2": `{"data":[{"symbol":"GOOG","name":"Alphabet","price":140,"logo_url":"u"},{"symbol":"AMZN","name":"Amazon","price":180,"logo_url":"u"}],"pages":3,"current_page":2,"count":5}`,
		"3": `{"data":[{"symbol":"NVDA","name":"NVIDIA","price":900,"logo_url":"u"}],"pages":3,"current_page":3,"count":5}`,
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
	for a, err := range c.Assets.ListAll(context.Background(), &assets.ListRequest{Limit: &limit}) {
		if err != nil {
			t.Fatalf("unexpected iteration error: %v", err)
		}
		got = append(got, a.Symbol)
	}
	want := []string{"AAPL", "MSFT", "GOOG", "AMZN", "NVDA"}
	if len(got) != len(want) {
		t.Fatalf("expected %d assets, got %d (%v)", len(want), len(got), got)
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
		_, _ = w.Write([]byte(`{"data":[{"symbol":"AAPL","name":"Apple","price":175,"logo_url":"u"}],"pages":10,"current_page":1,"count":10}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, err := range c.Assets.ListAll(context.Background(), nil) {
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
	for _, iterErr := range c.Assets.ListAll(context.Background(), nil) {
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

	_, err = c.Assets.List(context.Background(), nil)
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
