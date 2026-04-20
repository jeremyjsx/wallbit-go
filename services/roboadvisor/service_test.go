package roboadvisor_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/services/roboadvisor"
	"github.com/jeremyjsx/wallbit-go/wallbit"
)

func TestServiceGetBalance(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/roboadvisor/balance" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":1,"label":"Main Portfolio","category":null,"portfolio_type":"ROBOADVISOR","balance":1500,"portfolio_value":1450,"cash":50,"cash_available_withdrawal":50,"risk_profile":{"risk_level":3,"name":"Aggressive"},"performance":{"net_deposits":1000,"net_profits":500,"total_deposits":1200,"total_withdrawals":200},"assets":[{"symbol":"VTI","shares":5.2345,"market_value":1200,"price":229.18,"daily_variation_percentage":0.35,"weight":80,"logo":"https://static.atomicvest.com/VTI.svg"}],"allocation":{"cash":3.33,"securities":96.67},"has_pending_transactions":false}]}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.RoboAdvisor.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Payload.Data) != 1 {
		t.Fatalf("expected one portfolio, got %d", len(out.Payload.Data))
	}
	if out.Payload.Data[0].PortfolioType != "ROBOADVISOR" {
		t.Fatalf("unexpected portfolio_type %q", out.Payload.Data[0].PortfolioType)
	}
	if out.Payload.Data[0].RiskProfile == nil || out.Payload.Data[0].RiskProfile.RiskLevel != 3 {
		t.Fatalf("unexpected risk profile: %+v", out.Payload.Data[0].RiskProfile)
	}
}

func TestServiceGetBalanceNullRiskProfile(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":2,"label":null,"category":null,"portfolio_type":"CHEST","balance":100,"portfolio_value":0,"cash":100,"cash_available_withdrawal":100,"risk_profile":null,"performance":{"net_deposits":0,"net_profits":0,"total_deposits":0,"total_withdrawals":0},"assets":[],"allocation":{"cash":100,"securities":0},"has_pending_transactions":false}]}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.RoboAdvisor.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Payload.Data) != 1 {
		t.Fatalf("expected one portfolio, got %d", len(out.Payload.Data))
	}
	if out.Payload.Data[0].RiskProfile != nil {
		t.Fatalf("expected nil risk_profile, got %+v", out.Payload.Data[0].RiskProfile)
	}
}

func TestServiceGetBalanceReturnsAPIError(t *testing.T) {
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

	_, err = c.RoboAdvisor.GetBalance(context.Background())
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}

func TestServiceDeposit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/roboadvisor/deposit" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}

		var in roboadvisor.DepositRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.RoboAdvisorID != 1 || in.Amount != 500 || in.From != roboadvisor.AccountTypeDefault {
			t.Fatalf("unexpected payload: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"550e8400-e29b-41d4-a716-446655440000","type":"ROBOADVISOR_DEPOSIT","amount":500,"status":"PENDING","created_at":"2024-01-15T10:30:00+00:00"}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.RoboAdvisor.Deposit(context.Background(), roboadvisor.DepositRequest{
		RoboAdvisorID: 1,
		Amount:        500,
		From:          roboadvisor.AccountTypeDefault,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Type != "ROBOADVISOR_DEPOSIT" {
		t.Fatalf("unexpected transaction type %q", out.Payload.Data.Type)
	}
	if out.Payload.Data.Status != "PENDING" {
		t.Fatalf("unexpected status %q", out.Payload.Data.Status)
	}
}

func TestServiceDepositReturnsAPIError(t *testing.T) {
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

	_, err = c.RoboAdvisor.Deposit(context.Background(), roboadvisor.DepositRequest{
		RoboAdvisorID: 1,
		Amount:        500,
		From:          roboadvisor.AccountTypeDefault,
	})
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}

func TestServiceWithdraw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/roboadvisor/withdraw" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}

		var in roboadvisor.WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if in.RoboAdvisorID != 1 || in.Amount != 200 || in.To != roboadvisor.AccountTypeInvestment {
			t.Fatalf("unexpected payload: %+v", in)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"uuid":"660e9500-f30c-52e5-b827-557766551111","type":"ROBOADVISOR_WITHDRAW","amount":200,"status":"PENDING","created_at":"2024-01-15T14:00:00+00:00"}}`))
	}))
	defer server.Close()

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.RoboAdvisor.Withdraw(context.Background(), roboadvisor.WithdrawRequest{
		RoboAdvisorID: 1,
		Amount:        200,
		To:            roboadvisor.AccountTypeInvestment,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.Type != "ROBOADVISOR_WITHDRAW" {
		t.Fatalf("unexpected transaction type %q", out.Payload.Data.Type)
	}
	if out.Payload.Data.Status != "PENDING" {
		t.Fatalf("unexpected status %q", out.Payload.Data.Status)
	}
}

func TestServiceWithdrawReturnsAPIError(t *testing.T) {
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

	_, err = c.RoboAdvisor.Withdraw(context.Background(), roboadvisor.WithdrawRequest{
		RoboAdvisorID: 1,
		Amount:        200,
		To:            roboadvisor.AccountTypeInvestment,
	})
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
