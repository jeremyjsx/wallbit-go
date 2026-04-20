package cards_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/services/cards"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceBlockRejectsEmptyCardUUID(t *testing.T) {
	t.Parallel()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL("http://127.0.0.1:9"), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Cards.Block(context.Background(), "")
	if !errors.Is(err, cards.ErrEmptyCardUUID) {
		t.Fatalf("expected ErrEmptyCardUUID, got %v", err)
	}
	_, err = c.Cards.Unblock(context.Background(), "  ")
	if !errors.Is(err, cards.ErrEmptyCardUUID) {
		t.Fatalf("expected ErrEmptyCardUUID for whitespace, got %v", err)
	}
}

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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Payload.Data) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(out.Payload.Data))
	}
	if out.Payload.Data[0].UUID != "550e8400-e29b-41d4-a716-446655440000" || out.Payload.Data[0].CardLast4 != "1234" {
		t.Fatalf("unexpected first card: %+v", out.Payload.Data[0])
	}
	if out.Payload.Data[0].Expiration == nil || *out.Payload.Data[0].Expiration != "2029-01-01" {
		t.Fatalf("unexpected expiration: %v", out.Payload.Data[0].Expiration)
	}
	if out.Payload.Data[1].Expiration != nil {
		t.Fatalf("expected nil expiration, got %v", out.Payload.Data[1].Expiration)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Payload.Data) != 0 {
		t.Fatalf("expected empty list, got %d", len(out.Payload.Data))
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

	_, err = c.Cards.List(context.Background())
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.Block(context.Background(), cardUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.UUID != cardUUID {
		t.Fatalf("expected uuid %q, got %q", cardUUID, out.Payload.Data.UUID)
	}
	if out.Payload.Data.Status != cards.StatusSuspended {
		t.Fatalf("expected status %q, got %q", cards.StatusSuspended, out.Payload.Data.Status)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Cards.Block(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Cards.Unblock(context.Background(), cardUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.UUID != cardUUID {
		t.Fatalf("expected uuid %q, got %q", cardUUID, out.Payload.Data.UUID)
	}
	if out.Payload.Data.Status != cards.StatusActive {
		t.Fatalf("expected status %q, got %q", cards.StatusActive, out.Payload.Data.Status)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Cards.Unblock(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
