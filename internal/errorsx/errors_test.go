package errorsx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
)

func TestFromHTTPPreservesDetailsAsRawJSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"bad","code":"VALIDATION","details":{"field":["a"]}}`)
	err := errorsx.FromHTTP(http.StatusUnprocessableEntity, "req-1", body)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if len(sdkErr.Details) == 0 {
		t.Fatal("expected non-empty Details raw JSON")
	}
	var decoded struct {
		Field []string `json:"field"`
	}
	if err := json.Unmarshal(sdkErr.Details, &decoded); err != nil {
		t.Fatalf("unmarshal Details: %v", err)
	}
	if len(decoded.Field) != 1 || decoded.Field[0] != "a" {
		t.Fatalf("unexpected decoded details: %+v", decoded)
	}
}

func TestFromHTTPMessageAndDescription(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"primary","error":"secondary","code":"X"}`)
	err := errorsx.FromHTTP(http.StatusBadRequest, "", body)
	var e *errorsx.SDKError
	if !errors.As(err, &e) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if e.Message != "primary" {
		t.Fatalf("unexpected Message %q", e.Message)
	}
	if e.Description != "secondary" {
		t.Fatalf("unexpected Description %q", e.Description)
	}
}

func TestFromHTTPForbiddenYourPermissions(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"Insufficient permissions","your_permissions":["trade"]}`)
	err := errorsx.FromHTTP(http.StatusForbidden, "rid", body)
	var e *errorsx.SDKError
	if !errors.As(err, &e) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if len(e.YourPermissions) != 1 || e.YourPermissions[0] != "trade" {
		t.Fatalf("unexpected YourPermissions: %#v", e.YourPermissions)
	}
}

func TestFromHTTPRateLimitRetryAfter(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"slow down","retry_after":42}`)
	err := errorsx.FromHTTP(http.StatusTooManyRequests, "", body)
	var e *errorsx.SDKError
	if !errors.As(err, &e) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if e.RetryAfterSeconds == nil || *e.RetryAfterSeconds != 42 {
		t.Fatalf("unexpected RetryAfterSeconds: %v", e.RetryAfterSeconds)
	}
	if e.RetryAfter() != 42*time.Second {
		t.Fatalf("unexpected RetryAfter duration: %v", e.RetryAfter())
	}
}

func TestFromHTTPValidationErrors(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"invalid","errors":{"symbol":["required"]}}`)
	err := errorsx.FromHTTP(http.StatusUnprocessableEntity, "", body)
	var e *errorsx.SDKError
	if !errors.As(err, &e) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	var errs map[string][]string
	if err := json.Unmarshal(e.Errors, &errs); err != nil {
		t.Fatalf("unmarshal Errors: %v", err)
	}
	if errs["symbol"][0] != "required" {
		t.Fatalf("unexpected errors: %#v", errs)
	}
}

func TestPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		fn   func(error) bool
		want bool
	}{
		{"not found", errorsx.FromHTTP(http.StatusNotFound, "", nil), errorsx.IsNotFound, true},
		{"not found false for 400", errorsx.FromHTTP(http.StatusBadRequest, "", nil), errorsx.IsNotFound, false},
		{"auth 401", errorsx.FromHTTP(http.StatusUnauthorized, "", nil), errorsx.IsAuthError, true},
		{"auth 403", errorsx.FromHTTP(http.StatusForbidden, "", nil), errorsx.IsAuthError, true},
		{"rate limit", errorsx.FromHTTP(http.StatusTooManyRequests, "", nil), errorsx.IsRateLimit, true},
		{"validation 422", errorsx.FromHTTP(http.StatusUnprocessableEntity, "", nil), errorsx.IsValidationError, true},
		{"validation 400", errorsx.FromHTTP(http.StatusBadRequest, "", nil), errorsx.IsValidationError, true},
		{"server 500", errorsx.FromHTTP(http.StatusInternalServerError, "", nil), errorsx.IsServerError, true},
		{"server false for 404", errorsx.FromHTTP(http.StatusNotFound, "", nil), errorsx.IsServerError, false},
		{"retryable 429", errorsx.FromHTTP(http.StatusTooManyRequests, "", nil), errorsx.IsRetryable, true},
		{"retryable 503", errorsx.FromHTTP(http.StatusServiceUnavailable, "", nil), errorsx.IsRetryable, true},
		{"retryable false 400", errorsx.FromHTTP(http.StatusBadRequest, "", nil), errorsx.IsRetryable, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.fn(tt.err); got != tt.want {
				t.Fatalf("predicate: got %v want %v for %v", got, tt.want, tt.err)
			}
		})
	}
}

func TestIsTemporary(t *testing.T) {
	t.Parallel()

	e := errorsx.FromHTTP(http.StatusTooManyRequests, "", nil)
	if !e.IsTemporary() {
		t.Fatal("expected 429 temporary")
	}
	e2 := errorsx.FromHTTP(http.StatusBadGateway, "", nil)
	if !e2.IsTemporary() {
		t.Fatal("expected 502 temporary")
	}
	e3 := errorsx.FromHTTP(http.StatusBadRequest, "", nil)
	if e3.IsTemporary() {
		t.Fatal("expected 400 not temporary")
	}
}
