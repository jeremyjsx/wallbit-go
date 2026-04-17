package accountdetails

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const getPath = "/api/public/v1/account-details"

const (
	CountryUS = "US"
	CountryEU = "EU"

	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

type Service struct {
	sender httpx.Sender
}

func NewService(sender httpx.Sender) *Service {
	return &Service{sender: sender}
}

type GetRequest struct {
	Country  string
	Currency string
}

type AccountAddress struct {
	StreetLine1 string  `json:"street_line_1"`
	StreetLine2 *string `json:"street_line_2,omitempty"`
	City        string  `json:"city"`
	State       *string `json:"state,omitempty"`
	PostalCode  string  `json:"postal_code"`
	Country     string  `json:"country"`
}

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

type GetResponse struct {
	Data AccountDetails `json:"data"`
}

func (s *Service) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
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

	out := &GetResponse{}
	if err := s.sender.Send(ctx, http.MethodGet, path, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
