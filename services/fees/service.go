package fees

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const getPath = "/api/public/v1/fees"

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type GetRequest struct {
	Type string `json:"type"`
}

type FeeSetting struct {
	FeeType       string  `json:"fee_type"`
	Tier          *string `json:"tier"`
	PercentageFee string  `json:"percentage_fee"`
	FixedFeeUSD   string  `json:"fixed_fee_usd"`
}

type GetData struct {
	Row   *FeeSetting
	Empty bool
}

func (d *GetData) UnmarshalJSON(b []byte) error {
	*d = GetData{}
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return errors.New("fees: empty data")
	}
	switch b[0] {
	case '[':
		var arr []json.RawMessage
		if err := json.Unmarshal(b, &arr); err != nil {
			return err
		}
		if len(arr) != 0 {
			return fmt.Errorf("fees: data must be an empty array or an object, got %d array elements", len(arr))
		}
		d.Empty = true
		return nil
	case '{':
		var row FeeSetting
		if err := json.Unmarshal(b, &row); err != nil {
			return err
		}
		d.Row = &row
		return nil
	default:
		return fmt.Errorf("fees: unexpected data JSON")
	}
}

type GetResponse struct {
	Data GetData `json:"data"`
}

func (s *Service) Get(ctx context.Context, req GetRequest) (*transport.Response[GetResponse], error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := &GetResponse{}
	meta, err := s.sender.Send(ctx, http.MethodPost, getPath, bytes.NewBuffer(payload), out)
	if err != nil {
		return nil, err
	}

	return transport.NewResponse(meta, out), nil
}
