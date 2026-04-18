// Package errorsx turns Wallbit HTTP error responses into structured [SDKError] values.
//
// Retry guidance: treat an error as eligible for backoff-and-retry when [SDKError.IsTemporary]
// is true (HTTP 429 or any 5xx). Prefer honoring [SDKError.RetryAfterSeconds] when the API
// sends retry_after (Too Many Requests). Use exponential backoff with jitter; cap total wait
// and respect context cancellation. Do not retry 4xx other than 429 unless your application
// explicitly allows it.
package errorsx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// SDKError is returned for Wallbit API responses with status >= 400.
// Populated fields depend on the JSON body; unknown keys remain in [RawBody].
type SDKError struct {
	StatusCode int
	Code       string
	Message    string
	// Description is the API's optional "error" string (distinct from Message when both are set).
	Description string

	Details         json.RawMessage
	Errors          json.RawMessage
	YourPermissions []string

	// RetryAfterSeconds is set when the API includes retry_after (e.g. rate limit responses).
	RetryAfterSeconds *int64

	RequestID string
	RawBody   string
}

func (e *SDKError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("wallbit api error (%d:%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("wallbit api error (%d): %s", e.StatusCode, e.Message)
}

// IsTemporary reports whether the caller should consider backing off and retrying.
// True for HTTP 429 and any 5xx response.
func (e *SDKError) IsTemporary() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

// RetryAfter returns the server-suggested wait before retrying, or 0 if none was provided.
func (e *SDKError) RetryAfter() time.Duration {
	if e == nil || e.RetryAfterSeconds == nil {
		return 0
	}
	if *e.RetryAfterSeconds < 0 {
		return 0
	}
	return time.Duration(*e.RetryAfterSeconds) * time.Second
}

func asSDK(err error) (*SDKError, bool) {
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		return nil, false
	}
	return sdkErr, true
}

// IsNotFound reports whether err is an [SDKError] with HTTP 404.
func IsNotFound(err error) bool {
	e, ok := asSDK(err)
	return ok && e.StatusCode == http.StatusNotFound
}

// IsAuthError reports whether err is an [SDKError] with HTTP 401 or 403.
func IsAuthError(err error) bool {
	e, ok := asSDK(err)
	if !ok {
		return false
	}
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsRateLimit reports whether err is an [SDKError] with HTTP 429.
func IsRateLimit(err error) bool {
	e, ok := asSDK(err)
	return ok && e.StatusCode == http.StatusTooManyRequests
}

// IsValidationError reports whether err is an [SDKError] with HTTP 400 or 422.
func IsValidationError(err error) bool {
	e, ok := asSDK(err)
	if !ok {
		return false
	}
	return e.StatusCode == http.StatusUnprocessableEntity || e.StatusCode == http.StatusBadRequest
}

// IsServerError reports whether err is an [SDKError] with HTTP 5xx.
func IsServerError(err error) bool {
	e, ok := asSDK(err)
	return ok && e.StatusCode >= 500 && e.StatusCode <= 599
}

// IsRetryable reports whether err is an [SDKError] that is typically safe to retry with backoff
// (same condition as [SDKError.IsTemporary]: 429 or 5xx).
func IsRetryable(err error) bool {
	e, ok := asSDK(err)
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

func FromHTTP(statusCode int, requestID string, rawBody []byte) *SDKError {
	sdkErr := &SDKError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
		RequestID:  requestID,
		RawBody:    string(rawBody),
	}

	if len(rawBody) == 0 {
		return sdkErr
	}

	var body apiErrorBody
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return sdkErr
	}

	if body.Code != "" {
		sdkErr.Code = body.Code
	}
	if body.Message != "" {
		sdkErr.Message = body.Message
	} else if body.ErrorField != "" {
		sdkErr.Message = body.ErrorField
	}
	if body.ErrorField != "" {
		sdkErr.Description = body.ErrorField
	}
	sdkErr.Details = body.Details
	sdkErr.Errors = body.Errors
	if body.YourPermissions != nil {
		sdkErr.YourPermissions = append([]string(nil), body.YourPermissions...)
	}
	if body.RetryAfter != nil {
		sdkErr.RetryAfterSeconds = body.RetryAfter
	}

	return sdkErr
}
