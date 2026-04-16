package cards_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/cards"
)

func TestServiceBlock(t *testing.T) {
	t.Parallel()

	cardUUID := "550e8400-e29b-41d4-a716-446655440000"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/cards/"+cardUUID+"/status" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var in map[string]string
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in["status"] != cards.StatusSuspended {
			t.Fatalf("expected status %q, got %q", cards.StatusSuspended, in["status"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"` + cardUUID + `","status":"SUSPENDED"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.Block(context.Background(), cardUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.UUID != cardUUID {
		t.Fatalf("expected uuid %q, got %q", cardUUID, out.Data.UUID)
	}
	if out.Data.Status != cards.StatusSuspended {
		t.Fatalf("expected status %q, got %q", cards.StatusSuspended, out.Data.Status)
	}
}

func TestServiceBlockReturnsAPIError(t *testing.T) {
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

	_, err = c.Cards.Block(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}

func TestServiceUnblock(t *testing.T) {
	t.Parallel()

	cardUUID := "550e8400-e29b-41d4-a716-446655440000"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/cards/"+cardUUID+"/status" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var in map[string]string
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in["status"] != cards.StatusActive {
			t.Fatalf("expected status %q, got %q", cards.StatusActive, in["status"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"` + cardUUID + `","status":"ACTIVE"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.Unblock(context.Background(), cardUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.UUID != cardUUID {
		t.Fatalf("expected uuid %q, got %q", cardUUID, out.Data.UUID)
	}
	if out.Data.Status != cards.StatusActive {
		t.Fatalf("expected status %q, got %q", cards.StatusActive, out.Data.Status)
	}
}

func TestServiceUnblockReturnsAPIError(t *testing.T) {
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

	_, err = c.Cards.Unblock(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
