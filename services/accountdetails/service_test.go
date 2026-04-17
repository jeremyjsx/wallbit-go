package accountdetails_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/client"
	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/accountdetails"
)

func TestServiceGet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/v1/account-details" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if raw := r.URL.RawQuery; raw != "" {
			t.Fatalf("expected no query params, got %q", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"bank_name":"Community Federal Savings Bank","currency":"USD","account_type":"CHECKING","account_number":"9876543210","routing_number":"026073150","swift_code":"CTFNUS33","holder_name":"John Doe","address":{"street_line_1":"123 Main St","city":"New York","state":"NY","postal_code":"10001","country":"US"}}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.HolderName != "John Doe" {
		t.Fatalf("unexpected holder_name %q", out.Data.HolderName)
	}
	if out.Data.AccountType != "CHECKING" {
		t.Fatalf("unexpected account_type %q", out.Data.AccountType)
	}
	if out.Data.BankName != "Community Federal Savings Bank" {
		t.Fatalf("unexpected bank_name %q", out.Data.BankName)
	}
	if out.Data.Address == nil || out.Data.Address.City != "New York" {
		t.Fatalf("unexpected address: %+v", out.Data.Address)
	}
}

func TestServiceGetWithCountryAndCurrencyQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("country") != accountdetails.CountryEU {
			t.Fatalf("expected country=EU, got %q", r.URL.Query().Get("country"))
		}
		if r.URL.Query().Get("currency") != accountdetails.CurrencyEUR {
			t.Fatalf("expected currency=EUR, got %q", r.URL.Query().Get("currency"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"bank_name":"Deutsche Bank","currency":"EUR","account_type":"CHECKING","iban":"DE89370400440532013000","bic":"COBADEFFXXX","holder_name":"John Doe","address":{"street_line_1":"Hauptstraße 1","city":"Berlin","postal_code":"10115","country":"DE"}}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background(), &accountdetails.GetRequest{
		Country:  accountdetails.CountryEU,
		Currency: accountdetails.CurrencyEUR,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.IBAN == nil || *out.Data.IBAN != "DE89370400440532013000" {
		t.Fatalf("unexpected iban: %v", out.Data.IBAN)
	}
	if out.Data.BIC == nil || *out.Data.BIC != "COBADEFFXXX" {
		t.Fatalf("unexpected bic: %v", out.Data.BIC)
	}
}

func TestServiceGetForwardsArbitraryQueryValues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("country") != "XX" {
			t.Fatalf("expected country=XX forwarded, got %q", r.URL.Query().Get("country"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"bank_name":"X","currency":"USD","account_type":"CHECKING","holder_name":"H"}}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background(), &accountdetails.GetRequest{Country: "XX"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Data.HolderName != "H" {
		t.Fatalf("unexpected holder_name %q", out.Data.HolderName)
	}
}

func TestServiceGetReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden","code":"INSUFFICIENT_PERMISSIONS"}`))
	}))
	defer server.Close()

	c, err := client.NewClient("test-key", client.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.AccountDetails.Get(context.Background(), nil)
	var sdkErr *errorsx.SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %v", err)
	}
	if sdkErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", sdkErr.Code)
	}
}
