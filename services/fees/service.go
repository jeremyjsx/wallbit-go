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

// Service issues requests against the Wallbit fees endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// GetRequest selects which fee schedule to fetch. Type is the fee type key
// documented by the API (e.g. "card_issuance", "withdrawal_local").
type GetRequest struct {
	Type string `json:"type"`
}

// FeeSetting is the single fee row returned by the API. PercentageFee and
// FixedFeeUSD are kept as strings because the API serializes them that way
// (they are decimals and preserving the server-side precision matters).
type FeeSetting struct {
	FeeType       string  `json:"fee_type"`
	Tier          *string `json:"tier"`
	PercentageFee string  `json:"percentage_fee"`
	FixedFeeUSD   string  `json:"fixed_fee_usd"`
}

// GetData is the union payload for the fees endpoint. The API returns either
// an object (a concrete [FeeSetting], decoded into Row) or an empty array
// (Empty == true) when no rule applies to the requested fee type.
type GetData struct {
	Row   *FeeSetting
	Empty bool
}

// UnmarshalJSON decodes either a FeeSetting object or an empty JSON array
// into the receiver, rejecting non-empty arrays and any other shape.
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

// GetResponse is the top-level envelope for [Service.Get].
type GetResponse struct {
	Data GetData `json:"data"`
}

// Get fetches the fee schedule for req.Type. The response's Data may be
// either a concrete [FeeSetting] (when a rule exists) or an empty payload
// (when none does); consult [GetData] for the discriminator.
func (s *Service) Get(ctx context.Context, req GetRequest) (*transport.Response[GetResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodPost, getPath, req, &GetResponse{})
}
