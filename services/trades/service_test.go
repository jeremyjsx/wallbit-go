package trades_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/services/trades"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/trades" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var in trades.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.Symbol != "AAPL" || in.Direction != "BUY" || in.OrderType != "MARKET" {
			t.Fatalf("unexpected payload: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"symbol":"AAPL","direction":"BUY","amount":100,"shares":0.5,"status":"REQUESTED","order_type":"MARKET","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	amount := 100.0
	out, err := c.Trades.Create(context.Background(), trades.CreateRequest{
		Symbol:    "AAPL",
		Direction: "BUY",
		Currency:  "USD",
		OrderType: "MARKET",
		Amount:    &amount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Symbol != "AAPL" {
		t.Fatalf("expected symbol AAPL, got %q", out.Payload.Data.Symbol)
	}
	if out.Payload.Data.Status != "REQUESTED" {
		t.Fatalf("expected status REQUESTED, got %q", out.Payload.Data.Status)
	}
}

func TestServiceCreateLimitWithTimeInForce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in trades.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.OrderType != "LIMIT" {
			t.Fatalf("expected LIMIT, got %q", in.OrderType)
		}
		if in.TimeInForce == nil || *in.TimeInForce != "DAY" {
			t.Fatalf("expected time_in_force DAY, got %v", in.TimeInForce)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"symbol":"AAPL","direction":"BUY","amount":0,"shares":1,"status":"REQUESTED","order_type":"LIMIT","limit_price":150,"time_in_force":"DAY","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	limit := 150.0
	tif := "DAY"
	_, err = c.Trades.Create(context.Background(), trades.CreateRequest{
		Symbol:      "AAPL",
		Direction:   "BUY",
		Currency:    "USD",
		OrderType:   "LIMIT",
		Shares:      ptrFloat(1),
		LimitPrice:  &limit,
		TimeInForce: &tif,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ptrFloat(v float64) *float64 {
	return &v
}

func TestServiceCreateReturnsAPIError(t *testing.T) {
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

	amount := 100.0
	_, err = c.Trades.Create(context.Background(), trades.CreateRequest{
		Symbol:    "AAPL",
		Direction: "BUY",
		Currency:  "USD",
		OrderType: "MARKET",
		Amount:    &amount,
	})
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
