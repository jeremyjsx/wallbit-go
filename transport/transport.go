package transport

import (
	"context"
	"io"
	"net/http"
)

// Metadata describes the response envelope that every HTTP call returns,
// independent of the decoded payload. It is what [Sender] implementations
// report back to the per-resource services so they can expose it to end
// users through [Response].
//
// Header is the raw response header map (owned by the HTTP response and
// safe to read; callers should not mutate it). RequestID is the
// server-assigned identifier taken from the X-Request-ID header when
// present, empty otherwise. It is useful for support/debugging round-trips.
type Metadata struct {
	StatusCode int
	Header     http.Header
	RequestID  string
}

// Response is the generic wrapper returned by every service method. It
// pairs the HTTP response metadata with the typed, decoded payload T so
// callers can inspect both without giving up type safety.
//
// Payload is nil only when the endpoint legitimately returns no body
// (for example, an HTTP 204 No Content). For the common success case
// Payload is the fully decoded response struct.
type Response[T any] struct {
	StatusCode int
	Header     http.Header
	RequestID  string
	Payload    *T
}

// NewResponse assembles a [*Response] from the [*Metadata] returned by
// [Sender.Send] and a typed, already-decoded payload. It is a convenience
// for service implementations and nil-safe on meta so callers may use it
// without branching.
func NewResponse[T any](meta *Metadata, payload *T) *Response[T] {
	if meta == nil {
		return &Response[T]{Payload: payload}
	}
	return &Response[T]{
		StatusCode: meta.StatusCode,
		Header:     meta.Header,
		RequestID:  meta.RequestID,
		Payload:    payload,
	}
}

// Sender issues an authenticated HTTP request against the Wallbit API and
// decodes a JSON response into dest when non-nil. It is the seam used by
// the per-resource service packages to talk to the underlying client, and
// the extension point for tests, mocking, and middlewares (logging,
// metrics, distributed tracing, custom retries).
//
// Implementations must be safe for concurrent use by multiple goroutines.
// path is interpreted relative to the client's base URL. body may be nil
// for requests without a payload. dest may be nil for responses whose
// body is not consumed; otherwise it must be a non-nil pointer suitable
// for JSON unmarshalling.
//
// On success Send returns a non-nil [*Metadata] describing the final HTTP
// response. On transport failure (network error, context cancellation)
// Send returns a nil Metadata and a non-nil error. On an API error (HTTP
// status >= 400) Send returns a non-nil Metadata alongside the typed
// error so callers can surface status/headers/request-id even on failure.
type Sender interface {
	Send(ctx context.Context, method string, path string, body io.Reader, dest any) (*Metadata, error)
}
