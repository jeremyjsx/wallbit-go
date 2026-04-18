package errorsx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type SDKError struct {
	StatusCode int
	Code       string
	Message    string
	Details    json.RawMessage
	RequestID  string
	RawBody    string
}

func (e *SDKError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("wallbit api error (%d:%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("wallbit api error (%d): %s", e.StatusCode, e.Message)
}

func (e *SDKError) IsTemporary() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func IsAuthError(err error) bool {
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		return false
	}
	return sdkErr.StatusCode == http.StatusUnauthorized || sdkErr.StatusCode == http.StatusForbidden
}

func IsRateLimit(err error) bool {
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		return false
	}
	return sdkErr.StatusCode == http.StatusTooManyRequests
}

func IsValidationError(err error) bool {
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		return false
	}
	return sdkErr.StatusCode == http.StatusUnprocessableEntity || sdkErr.StatusCode == http.StatusBadRequest
}

type apiErrorEnvelope struct {
	Error   string          `json:"error"`
	Message string          `json:"message"`
	Code    string          `json:"code"`
	Details json.RawMessage `json:"details"`
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

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return sdkErr
	}

	if envelope.Code != "" {
		sdkErr.Code = envelope.Code
	}
	if envelope.Message != "" {
		sdkErr.Message = envelope.Message
	}
	if envelope.Error != "" && sdkErr.Message == http.StatusText(statusCode) {
		sdkErr.Message = envelope.Error
	}
	sdkErr.Details = envelope.Details

	return sdkErr
}
