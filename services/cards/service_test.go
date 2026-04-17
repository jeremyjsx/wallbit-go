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

func TestServiceList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/cards" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query string, got %q", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"uuid":"550e8400-e29b-41d4-a716-446655440000","status":"ACTIVE","card_type":"VIRTUAL","card_network":"visa","card_last4":"1234","expiration":"2029-01-01"},{"uuid":"c37ea154-f6d2-4c95-b3f4-80f07858db7f","status":"SUSPENDED","card_type":"PHYSICAL","card_network":"mastercard","card_last4":"5678","expiration":null}]}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(out.Data))
	}
	if out.Data[0].UUID != "550e8400-e29b-41d4-a716-446655440000" || out.Data[0].CardLast4 != "1234" {
		t.Fatalf("unexpected first card: %+v", out.Data[0])
	}
	if out.Data[0].Expiration == nil || *out.Data[0].Expiration != "2029-01-01" {
		t.Fatalf("unexpected expiration: %v", out.Data[0].Expiration)
	}
	if out.Data[1].Expiration != nil {
		t.Fatalf("expected nil expiration, got %v", out.Data[1].Expiration)
	}
}

func TestServiceListEmpty(t *testing.T) {
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

	out, err := c.Cards.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Data) != 0 {
		t.Fatalf("expected empty list, got %d", len(out.Data))
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

	_, err = c.Cards.List(context.Background())
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}

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
