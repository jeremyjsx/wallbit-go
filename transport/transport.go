package transport

import (
	"context"
	"io"
)

// Sender issues an authenticated HTTP request against the Wallbit API and
// decodes a JSON response into dest when non-nil. It is the seam used by the
// per-resource service packages to talk to the underlying client, and the
// extension point for tests, mocking, and middlewares (logging, metrics,
// distributed tracing, custom retries).
//
// Implementations must be safe for concurrent use by multiple goroutines.
// path is interpreted relative to the client's base URL. body may be nil for
// requests without a payload. dest may be nil for responses whose body is
// not consumed; otherwise it must be a non-nil pointer suitable for JSON
// unmarshalling.
type Sender interface {
	Send(ctx context.Context, method string, path string, body io.Reader, dest any) error
}
