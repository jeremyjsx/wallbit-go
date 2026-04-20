package wallbit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Error is returned for Wallbit API responses with HTTP status >= 400.
// Populated fields depend on the JSON body shape; unknown keys remain in
// [Error.RawBody]. Use [errors.As] to extract the typed error from a wrapped
// chain, and the IsX predicates ([IsNotFound], [IsAuthError], [IsRateLimit],
// [IsValidationError], [IsServerError], [IsRetryable]) to branch on
// categories.
//
// Retry guidance: treat an error as eligible for backoff-and-retry when
// [Error.IsTemporary] is true (HTTP 429 or any 5xx). Prefer honoring
// [Error.RetryAfter] when set. Use exponential backoff with jitter, cap
// total wait, and respect context cancellation. Do not retry 4xx other
// than 429 unless your application explicitly allows it.
type Error struct {
	StatusCode int
	Code       string
	Message    string
	// Description is the API's optional "error" string (distinct from
	// Message when both are set).
	Description string

	Details         json.RawMessage
	Errors          json.RawMessage
	YourPermissions []string

	// RetryAfterSeconds is set when the API includes retry_after (e.g.
	// rate limit responses).
	RetryAfterSeconds *int64

	RequestID string
	RawBody   string
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("wallbit api error (%d:%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("wallbit api error (%d): %s", e.StatusCode, e.Message)
}

// IsTemporary reports whether the caller should consider backing off and
// retrying. True for HTTP 429 and any 5xx response.
func (e *Error) IsTemporary() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

// RetryAfter returns the server-suggested wait before retrying, or 0 if
// none was provided.
func (e *Error) RetryAfter() time.Duration {
	if e == nil || e.RetryAfterSeconds == nil {
		return 0
	}
	if *e.RetryAfterSeconds < 0 {
		return 0
	}
	return time.Duration(*e.RetryAfterSeconds) * time.Second
}

func asAPIError(err error) (*Error, bool) {
	var e *Error
	if !errors.As(err, &e) {
		return nil, false
	}
	return e, true
}

// IsNotFound reports whether err is an [Error] with HTTP 404.
func IsNotFound(err error) bool {
	e, ok := asAPIError(err)
	return ok && e.StatusCode == http.StatusNotFound
}

// IsAuthError reports whether err is an [Error] with HTTP 401 or 403.
func IsAuthError(err error) bool {
	e, ok := asAPIError(err)
	if !ok {
		return false
	}
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsRateLimit reports whether err is an [Error] with HTTP 429.
func IsRateLimit(err error) bool {
	e, ok := asAPIError(err)
	return ok && e.StatusCode == http.StatusTooManyRequests
}

// IsValidationError reports whether err is an [Error] with HTTP 400 or 422.
func IsValidationError(err error) bool {
	e, ok := asAPIError(err)
	if !ok {
		return false
	}
	return e.StatusCode == http.StatusUnprocessableEntity || e.StatusCode == http.StatusBadRequest
}

// IsServerError reports whether err is an [Error] with HTTP 5xx.
func IsServerError(err error) bool {
	e, ok := asAPIError(err)
	return ok && e.StatusCode >= 500 && e.StatusCode <= 599
}

// IsRetryable reports whether err is an [Error] that is typically safe to
// retry with backoff (HTTP 429 or 5xx).
func IsRetryable(err error) bool {
	e, ok := asAPIError(err)
	if !ok {
		return false
	}
	return e.IsTemporary()
}

type apiErrorBody struct {
	ErrorField      string          `json:"error"`
	Message         string          `json:"message"`
	Code            string          `json:"code"`
	Details         json.RawMessage `json:"details"`
	Errors          json.RawMessage `json:"errors"`
	YourPermissions []string        `json:"your_permissions"`
	RetryAfter      *int64          `json:"retry_after"`
}

// ErrorFromHTTP constructs an [Error] from a raw HTTP response. The JSON
// body is parsed best-effort; unknown shapes leave [Error.Message] set to
// the standard HTTP status text and the original payload available in
// [Error.RawBody]. Exposed primarily for tests and consumers that mock
// transports.
func ErrorFromHTTP(statusCode int, requestID string, rawBody []byte) *Error {
	apiErr := &Error{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
		RequestID:  requestID,
		RawBody:    string(rawBody),
	}

	if len(rawBody) == 0 {
		return apiErr
	}

	var body apiErrorBody
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return apiErr
	}

	if body.Code != "" {
		apiErr.Code = body.Code
	}
	if body.Message != "" {
		apiErr.Message = body.Message
	} else if body.ErrorField != "" {
		apiErr.Message = body.ErrorField
	}
	if body.ErrorField != "" {
		apiErr.Description = body.ErrorField
	}
	apiErr.Details = body.Details
	apiErr.Errors = body.Errors
	if body.YourPermissions != nil {
		apiErr.YourPermissions = append([]string(nil), body.YourPermissions...)
	}
	if body.RetryAfter != nil {
		apiErr.RetryAfterSeconds = body.RetryAfter
	}

	return apiErr
}
