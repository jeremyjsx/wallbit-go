package wallbit_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestErrorFromHTTPPreservesDetailsAsRawJSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"bad","code":"VALIDATION","details":{"field":["a"]}}`)
	err := wallbit.ErrorFromHTTP(http.StatusUnprocessableEntity, "req-1", body)
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if len(apiErr.Details) == 0 {
		t.Fatal("expected non-empty Details raw JSON")
	}
	var decoded struct {
		Field []string `json:"field"`
	}
	if err := json.Unmarshal(apiErr.Details, &decoded); err != nil {
		t.Fatalf("unmarshal Details: %v", err)
	}
	if len(decoded.Field) != 1 || decoded.Field[0] != "a" {
		t.Fatalf("unexpected decoded details: %+v", decoded)
	}
}

func TestErrorFromHTTPMessageAndDescription(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"primary","error":"secondary","code":"X"}`)
	err := wallbit.ErrorFromHTTP(http.StatusBadRequest, "", body)
	var e *wallbit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if e.Message != "primary" {
		t.Fatalf("unexpected Message %q", e.Message)
	}
	if e.Description != "secondary" {
		t.Fatalf("unexpected Description %q", e.Description)
	}
}

func TestErrorFromHTTPForbiddenYourPermissions(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"Insufficient permissions","your_permissions":["trade"]}`)
	err := wallbit.ErrorFromHTTP(http.StatusForbidden, "rid", body)
	var e *wallbit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if len(e.YourPermissions) != 1 || e.YourPermissions[0] != "trade" {
		t.Fatalf("unexpected YourPermissions: %#v", e.YourPermissions)
	}
}

func TestErrorFromHTTPRateLimitRetryAfter(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"slow down","retry_after":42}`)
	err := wallbit.ErrorFromHTTP(http.StatusTooManyRequests, "", body)
	var e *wallbit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if e.RetryAfterSeconds == nil || *e.RetryAfterSeconds != 42 {
		t.Fatalf("unexpected RetryAfterSeconds: %v", e.RetryAfterSeconds)
	}
	if e.RetryAfter() != 42*time.Second {
		t.Fatalf("unexpected RetryAfter duration: %v", e.RetryAfter())
	}
}

func TestErrorFromHTTPValidationErrors(t *testing.T) {
	t.Parallel()

	body := []byte(`{"message":"invalid","errors":{"symbol":["required"]}}`)
	err := wallbit.ErrorFromHTTP(http.StatusUnprocessableEntity, "", body)
	var e *wallbit.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	var errs map[string][]string
	if err := json.Unmarshal(e.Errors, &errs); err != nil {
		t.Fatalf("unmarshal Errors: %v", err)
	}
	if errs["symbol"][0] != "required" {
		t.Fatalf("unexpected errors: %#v", errs)
	}
}

func TestErrorPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		fn   func(error) bool
		want bool
	}{
		{"not found", wallbit.ErrorFromHTTP(http.StatusNotFound, "", nil), wallbit.IsNotFound, true},
		{"not found false for 400", wallbit.ErrorFromHTTP(http.StatusBadRequest, "", nil), wallbit.IsNotFound, false},
		{"auth 401", wallbit.ErrorFromHTTP(http.StatusUnauthorized, "", nil), wallbit.IsAuthError, true},
		{"auth 403", wallbit.ErrorFromHTTP(http.StatusForbidden, "", nil), wallbit.IsAuthError, true},
		{"rate limit", wallbit.ErrorFromHTTP(http.StatusTooManyRequests, "", nil), wallbit.IsRateLimit, true},
		{"validation 422", wallbit.ErrorFromHTTP(http.StatusUnprocessableEntity, "", nil), wallbit.IsValidationError, true},
		{"validation 400", wallbit.ErrorFromHTTP(http.StatusBadRequest, "", nil), wallbit.IsValidationError, true},
		{"server 500", wallbit.ErrorFromHTTP(http.StatusInternalServerError, "", nil), wallbit.IsServerError, true},
		{"server false for 404", wallbit.ErrorFromHTTP(http.StatusNotFound, "", nil), wallbit.IsServerError, false},
		{"retryable 429", wallbit.ErrorFromHTTP(http.StatusTooManyRequests, "", nil), wallbit.IsRetryable, true},
		{"retryable 503", wallbit.ErrorFromHTTP(http.StatusServiceUnavailable, "", nil), wallbit.IsRetryable, true},
		{"retryable false 400", wallbit.ErrorFromHTTP(http.StatusBadRequest, "", nil), wallbit.IsRetryable, false},
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

func TestErrorIsTemporary(t *testing.T) {
	t.Parallel()

	e := wallbit.ErrorFromHTTP(http.StatusTooManyRequests, "", nil)
	if !e.IsTemporary() {
		t.Fatal("expected 429 temporary")
	}
	e2 := wallbit.ErrorFromHTTP(http.StatusBadGateway, "", nil)
	if !e2.IsTemporary() {
		t.Fatal("expected 502 temporary")
	}
	e3 := wallbit.ErrorFromHTTP(http.StatusBadRequest, "", nil)
	if e3.IsTemporary() {
		t.Fatal("expected 400 not temporary")
	}
}
