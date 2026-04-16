package accountdetails_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/account-details" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"account_holder":"John Doe","bank_name":"Bank Inc.","routing_number":"123456789","account_number":"000123456789","currency":"USD","type":"checking","ach_type":"personal"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.AccountHolder != "John Doe" {
		t.Fatalf("unexpected account holder %q", out.Data.AccountHolder)
	}
	if out.Data.Type != "checking" {
		t.Fatalf("unexpected type %q", out.Data.Type)
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

	_, err = c.AccountDetails.Get(context.Background())
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
