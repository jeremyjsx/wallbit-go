package trades

import (
	"context"
	"net/http"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const createPath = "/api/public/v1/trades"

// Service issues requests against the Wallbit trades endpoint.
type Service struct {
	sender transport.Sender
}

// NewService wires a [Service] to the given [transport.Sender].
func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

// CreateRequest parameterises a call to [Service.Create]. Exactly one of
// Amount (notional in Currency) or Shares (quantity) must be provided; the
// other price fields depend on OrderType (e.g. LimitPrice for "LIMIT",
// StopPrice for "STOP"). See the Wallbit docs for the full combination
// matrix.
type CreateRequest struct {
	Symbol      string   `json:"symbol"`
	Direction   string   `json:"direction"`
	Currency    string   `json:"currency"`
	OrderType   string   `json:"order_type"`
	Amount      *float64 `json:"amount,omitempty"`
	Shares      *float64 `json:"shares,omitempty"`
	StopPrice   *float64 `json:"stop_price,omitempty"`
	LimitPrice  *float64 `json:"limit_price,omitempty"`
	TimeInForce *string  `json:"time_in_force,omitempty"`
}

// Trade is the order row returned by [Service.Create]. UpdatedAt tracks
// the last status transition; for terminal statuses it equals the fill
// timestamp.
type Trade struct {
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"`
	Amount      float64   `json:"amount"`
	Shares      float64   `json:"shares"`
	Status      string    `json:"status"`
	OrderType   string    `json:"order_type"`
	LimitPrice  *float64  `json:"limit_price"`
	StopPrice   *float64  `json:"stop_price"`
	TimeInForce *string   `json:"time_in_force"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateResponse is the top-level envelope for [Service.Create].
type CreateResponse struct {
	Data Trade `json:"data"`
}

// Create submits a new equity order for the symbol and order type
// specified in req.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*transport.Response[CreateResponse], error) {
	return transport.SendJSON(ctx, s.sender, http.MethodPost, createPath, req, &CreateResponse{})
}
