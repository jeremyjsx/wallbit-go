package accountdetails

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const getPath = "/api/public/v1/account-details"

type Service struct {
	sender httpx.Sender
}

func NewService(sender httpx.Sender) *Service {
	return &Service{sender: sender}
}

type AccountDetails struct {
	AccountHolder string         `json:"account_holder"`
	BankName      string         `json:"bank_name"`
	RoutingNumber string         `json:"routing_number"`
	AccountNumber string         `json:"account_number"`
	IBAN          string         `json:"iban"`
	SWIFTBIC      string         `json:"swift_bic"`
	Currency      string         `json:"currency"`
	Type          string         `json:"type"`
	Extra         map[string]any `json:"-"`
}

func (a *AccountDetails) UnmarshalJSON(data []byte) error {
	type Alias AccountDetails
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	knownFields := map[string]struct{}{
		"account_holder": {},
		"bank_name":      {},
		"routing_number": {},
		"account_number": {},
		"iban":           {},
		"swift_bic":      {},
		"currency":       {},
		"type":           {},
	}

	for k := range knownFields {
		delete(raw, k)
	}
	a.Extra = raw

	return nil
}

type GetResponse struct {
	Data AccountDetails `json:"data"`
}

func (s *Service) Get(ctx context.Context) (*GetResponse, error) {
	out := &GetResponse{}
	if err := s.sender.Send(ctx, http.MethodGet, getPath, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
