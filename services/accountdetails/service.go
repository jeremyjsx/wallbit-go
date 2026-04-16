package accountdetails

import (
	"context"
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
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	RoutingNumber string `json:"routing_number"`
	AccountNumber string `json:"account_number"`
	IBAN          string `json:"iban"`
	SWIFTBIC      string `json:"swift_bic"`
	Currency      string `json:"currency"`
	Type          string `json:"type"`
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
