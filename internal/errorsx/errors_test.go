package errorsx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

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
