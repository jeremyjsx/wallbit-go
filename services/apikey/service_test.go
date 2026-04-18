package apikey_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceRevoke(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/api-key" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"API Key revoked successfully."}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.APIKey.Revoke(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Message != "API Key revoked successfully." {
		t.Fatalf("unexpected message %q", out.Message)
	}
}

func TestServiceRevokeReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid or expired API Key."}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.APIKey.Revoke(context.Background())
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", sdkErr.StatusCode)
	}
}
