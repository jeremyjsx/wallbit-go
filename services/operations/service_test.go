package operations_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/operations"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceInternal(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/operations/internal" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var in operations.InternalRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.Currency != "USD" || in.From != operations.AccountDefault || in.To != operations.AccountInvestment || in.Amount != 100 {
			t.Fatalf("unexpected payload: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"tx_123","type":"INTERNAL_OPERATION","external_address":null,"source_currency":{"code":"USD","alias":"US Dollar"},"dest_currency":{"code":"USD","alias":"US Dollar"},"source_amount":100,"dest_amount":100,"status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","comment":null}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Operations.Internal(context.Background(), operations.InternalRequest{
		Currency: "USD",
		From:     operations.AccountDefault,
		To:       operations.AccountInvestment,
		Amount:   100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.UUID != "tx_123" {
		t.Fatalf("expected uuid tx_123, got %q", out.Data.UUID)
	}
	if out.Data.Status != "COMPLETED" {
		t.Fatalf("expected status COMPLETED, got %q", out.Data.Status)
	}
}

func TestServiceDepositInvestment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in operations.InternalRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.From != operations.AccountDefault || in.To != operations.AccountInvestment {
			t.Fatalf("unexpected account movement for deposit: %+v", in)
		}
		if in.Currency != "USD" || in.Amount != 25 {
			t.Fatalf("unexpected payload for deposit: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"tx_deposit","type":"INTERNAL_OPERATION","external_address":null,"source_currency":{"code":"USD","alias":"US Dollar"},"dest_currency":{"code":"USD","alias":"US Dollar"},"source_amount":25,"dest_amount":25,"status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","comment":null}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Operations.DepositInvestment(context.Background(), operations.InvestmentDepositRequest{
		Currency: "USD",
		Amount:   25,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.UUID != "tx_deposit" {
		t.Fatalf("expected uuid tx_deposit, got %q", out.Data.UUID)
	}
}

func TestServiceWithdrawInvestment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in operations.InternalRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.From != operations.AccountInvestment || in.To != operations.AccountDefault {
			t.Fatalf("unexpected account movement for withdrawal: %+v", in)
		}
		if in.Currency != "USD" || in.Amount != 15 {
			t.Fatalf("unexpected payload for withdrawal: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"tx_withdraw","type":"INTERNAL_OPERATION","external_address":null,"source_currency":{"code":"USD","alias":"US Dollar"},"dest_currency":{"code":"USD","alias":"US Dollar"},"source_amount":15,"dest_amount":15,"status":"COMPLETED","created_at":"2024-01-01T00:00:00Z","comment":null}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.Operations.WithdrawInvestment(context.Background(), operations.InvestmentWithdrawRequest{
		Currency: "USD",
		Amount:   15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.UUID != "tx_withdraw" {
		t.Fatalf("expected uuid tx_withdraw, got %q", out.Data.UUID)
	}
}

func TestServiceInternalReturnsAPIError(t *testing.T) {
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

	_, err = c.Operations.Internal(context.Background(), operations.InternalRequest{
		Currency: "USD",
		From:     operations.AccountDefault,
		To:       operations.AccountInvestment,
		Amount:   10,
	})
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
