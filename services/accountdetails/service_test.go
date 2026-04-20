package accountdetails_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeremyjsx/wallbit-go/services/accountdetails"
	"github.com/jeremyjsx/wallbit-go/wallbit"
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.HolderName != "John Doe" {
		t.Fatalf("unexpected holder_name %q", out.Payload.Data.HolderName)
	}
	if out.Payload.Data.AccountType != "CHECKING" {
		t.Fatalf("unexpected account_type %q", out.Payload.Data.AccountType)
	}
	if out.Payload.Data.BankName != "Community Federal Savings Bank" {
		t.Fatalf("unexpected bank_name %q", out.Payload.Data.BankName)
	}
	if out.Payload.Data.Address == nil || out.Payload.Data.Address.City != "New York" {
		t.Fatalf("unexpected address: %+v", out.Payload.Data.Address)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
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
	if out.Payload.Data.IBAN == nil || *out.Payload.Data.IBAN != "DE89370400440532013000" {
		t.Fatalf("unexpected iban: %v", out.Payload.Data.IBAN)
	}
	if out.Payload.Data.BIC == nil || *out.Payload.Data.BIC != "COBADEFFXXX" {
		t.Fatalf("unexpected bic: %v", out.Payload.Data.BIC)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := c.AccountDetails.Get(context.Background(), &accountdetails.GetRequest{Country: "XX"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Payload.Data.HolderName != "H" {
		t.Fatalf("unexpected holder_name %q", out.Payload.Data.HolderName)
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

	c, err := wallbit.NewClient("test-key", wallbit.WithBaseURL(server.URL), wallbit.WithInsecureHTTPForTesting())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.AccountDetails.Get(context.Background(), nil)
	var apiErr *wallbit.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *wallbit.Error, got %v", err)
	}
	if apiErr.Code != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("unexpected error code %q", apiErr.Code)
	}
}
