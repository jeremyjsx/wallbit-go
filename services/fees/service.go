package fees

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/internal/httpx"
)

const getPath = "/api/public/v1/fees"

type Service struct {
	sender httpx.Sender
}

func NewService(sender httpx.Sender) *Service {
	return &Service{sender: sender}
}

type GetRequest struct {
	Type string `json:"type"`
}

type FeeSetting struct {
	FeeType       string `json:"fee_type"`
	Tier          string `json:"tier"`
	PercentageFee string `json:"percentage_fee"`
	FixedFeeUSD   string `json:"fixed_fee_usd"`
}

type GetResponse struct {
	Data any `json:"data"`
}

func (s *Service) Get(ctx context.Context, req GetRequest) (*GetResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &GetResponse{}
	if err := s.sender.Send(ctx, http.MethodPost, getPath, bytes.NewBuffer(payload), out); err != nil {
		return nil, err
	}

	return out, nil
}
