package accountdetails

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const getPath = "/api/public/v1/account-details"

// Well-known country and currency codes accepted by the API. These are
// plain string constants because the endpoint may add new values over time;
// they exist purely as a documented starting point.
const (
	CountryUS = "US"
	CountryEU = "EU"

	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

// Service issues requests against the Wallbit account-details endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// GetRequest parameterises a call to [Service.Get]. Both fields are
// optional; when blank, the server returns the account details configured
// as default for the authenticated user.
type GetRequest struct {
	Country  string
	Currency string
}

// AccountAddress is the postal address attached to a set of [AccountDetails].
type AccountAddress struct {
	StreetLine1 string  `json:"street_line_1"`
	StreetLine2 *string `json:"street_line_2,omitempty"`
	City        string  `json:"city"`
	State       *string `json:"state,omitempty"`
	PostalCode  string  `json:"postal_code"`
	Country     string  `json:"country"`
}

// AccountDetails describes the bank account configured for the caller to
// deposit or withdraw fiat. Optional fields (IBAN, BIC, routing number, …)
// are nil when the destination bank network does not use them.
type AccountDetails struct {
	BankName      string          `json:"bank_name"`
	Currency      string          `json:"currency"`
	AccountType   string          `json:"account_type"`
	AccountNumber *string         `json:"account_number,omitempty"`
	RoutingNumber *string         `json:"routing_number,omitempty"`
	IBAN          *string         `json:"iban,omitempty"`
	BIC           *string         `json:"bic,omitempty"`
	SWIFTCode     *string         `json:"swift_code,omitempty"`
	HolderName    string          `json:"holder_name"`
	Beneficiary   *string         `json:"beneficiary,omitempty"`
	Memo          *string         `json:"memo,omitempty"`
	Address       *AccountAddress `json:"address,omitempty"`
}

// GetResponse is the top-level envelope for [Service.Get].
type GetResponse struct {
	Data AccountDetails `json:"data"`
}

// Get returns the bank account details the user should use to fund or
// withdraw from their Wallbit account. A nil req uses server defaults.
func (s *Service) Get(ctx context.Context, req *GetRequest) (*transport.Response[GetResponse], error) {
	path := getPath
	if req != nil {
		q := url.Values{}
		if req.Country != "" {
			q.Set("country", req.Country)
		}
		if req.Currency != "" {
			q.Set("currency", req.Currency)
		}
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &GetResponse{})
}
