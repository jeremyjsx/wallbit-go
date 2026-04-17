package fees_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/fees"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/fees" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var in fees.GetRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.Type != "TRADE" {
			t.Fatalf("expected type TRADE, got %q", in.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"fee_type":"TRADE","tier":"LEVEL1","percentage_fee":"0.005","fixed_fee_usd":"0.00"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Fees.Get(context.Background(), fees.GetRequest{Type: "TRADE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.Empty {
		t.Fatal("expected fee row, got empty data")
	}
	if out.Data.Row == nil {
		t.Fatal("expected non-nil fee setting")
	}
	if out.Data.Row.FeeType != "TRADE" {
		t.Fatalf("expected fee_type TRADE, got %q", out.Data.Row.FeeType)
	}
}

func TestServiceGetReturnsEmptyDataArray(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Fees.Get(context.Background(), fees.GetRequest{Type: "TRADE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Data.Empty {
		t.Fatalf("expected empty data flag, got %#v", out.Data)
	}
	if out.Data.Row != nil {
		t.Fatalf("expected nil row for empty data, got %+v", out.Data.Row)
	}
}

func TestServiceGetRejectsNonEmptyDataArray(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"fee_type":"TRADE"}]}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Fees.Get(context.Background(), fees.GetRequest{Type: "TRADE"})
	if err == nil {
		t.Fatal("expected error for non-empty data array")
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

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Fees.Get(context.Background(), fees.GetRequest{Type: "TRADE"})
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
