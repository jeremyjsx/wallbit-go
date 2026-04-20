package operations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const internalPath = "/api/public/v1/operations/internal"

const (
	AccountDefault    = "DEFAULT"
	AccountInvestment = "INVESTMENT"
)

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type InternalRequest struct {
	Currency string  `json:"currency"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
}

type InvestmentDepositRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type InvestmentWithdrawRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type Currency struct {
	Code  string `json:"code"`
	Alias string `json:"alias"`
}

type Transaction struct {
	UUID            string   `json:"uuid"`
	Type            string   `json:"type"`
	ExternalAddress *string  `json:"external_address"`
	SourceCurrency  Currency `json:"source_currency"`
	DestCurrency    Currency `json:"dest_currency"`
	SourceAmount    float64  `json:"source_amount"`
	DestAmount      float64  `json:"dest_amount"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
	Comment         *string  `json:"comment"`
}

type InternalResponse struct {
	Data Transaction `json:"data"`
}

func (s *Service) Internal(ctx context.Context, req InternalRequest) (*InternalResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &InternalResponse{}
	if err := s.sender.Send(ctx, http.MethodPost, internalPath, bytes.NewBuffer(payload), out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) DepositInvestment(ctx context.Context, req InvestmentDepositRequest) (*InternalResponse, error) {
	return s.Internal(ctx, InternalRequest{
		Currency: req.Currency,
		From:     AccountDefault,
		To:       AccountInvestment,
		Amount:   req.Amount,
	})
}

func (s *Service) WithdrawInvestment(ctx context.Context, req InvestmentWithdrawRequest) (*InternalResponse, error) {
	return s.Internal(ctx, InternalRequest{
		Currency: req.Currency,
		From:     AccountInvestment,
		To:       AccountDefault,
		Amount:   req.Amount,
	})
}
