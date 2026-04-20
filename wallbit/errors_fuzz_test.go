package wallbit_test

import (
	"strings"
	"testing"

	"github.com/jeremyjsx/wallbit-go/wallbit"
)

// FuzzErrorFromHTTP exercises the parser against arbitrary inputs to
// ensure no shape causes a panic, returns nil, or silently drops the
// verbatim fields the caller provided. The seed corpus covers the
// documented response shapes; the fuzzer mutates around them.
//
// Run extended fuzzing locally with:
//
//	go test -fuzz=FuzzErrorFromHTTP -fuzztime=30s ./wallbit
func FuzzErrorFromHTTP(f *testing.F) {
	bodies := [][]byte{
		nil,
		[]byte(``),
		[]byte(`{}`),
		[]byte(`{"message":"resource not found"}`),
		[]byte(`{"code":"INSUFFICIENT_PERMISSIONS","message":"forbidden"}`),
		[]byte(`{"retry_after":3,"message":"slow down"}`),
		[]byte(`{"retry_after":-7}`),
		[]byte(`{"details":{"field":"amount","reason":"too_small"}}`),
		[]byte(`{"errors":[{"field":"x"},{"field":"y"}]}`),
		[]byte(`{"your_permissions":["READ","WRITE"]}`),
		[]byte(`{"error":"trade rejected"}`),
		[]byte(`not json at all`),
		[]byte(`{"message":"` + strings.Repeat("x", 4096) + `"}`),
	}
	statuses := []int{200, 400, 401, 404, 422, 429, 500, 502, 0, -1}
	for _, b := range bodies {
		for _, s := range statuses {
			f.Add(s, "req-test", b)
		}
	}

	f.Fuzz(func(t *testing.T, statusCode int, requestID string, rawBody []byte) {
		e := wallbit.ErrorFromHTTP(statusCode, requestID, rawBody)
		if e == nil {
			t.Fatal("ErrorFromHTTP returned nil")
		}
		if e.StatusCode != statusCode {
			t.Fatalf("StatusCode: got %d, want %d", e.StatusCode, statusCode)
		}
		if e.RequestID != requestID {
			t.Fatalf("RequestID: got %q, want %q", e.RequestID, requestID)
		}
		if e.RawBody != string(rawBody) {
			t.Fatalf("RawBody not preserved")
		}
		if d := e.RetryAfter(); d < 0 {
			t.Fatalf("RetryAfter returned negative duration: %s", d)
		}
		// Methods and predicates must never panic, regardless of the
		// shape of the body or the status code (including non-HTTP
		// values from a buggy upstream proxy).
		_ = e.Error()
		_ = e.IsTemporary()
		_ = wallbit.IsNotFound(e)
		_ = wallbit.IsAuthError(e)
		_ = wallbit.IsRateLimit(e)
		_ = wallbit.IsValidationError(e)
		_ = wallbit.IsServerError(e)
		_ = wallbit.IsRetryable(e)
	})
}
