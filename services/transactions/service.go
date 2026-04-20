package transactions

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const listPath = "/api/public/v1/transactions"

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type ListRequest struct {
	Page       *int
	Limit      *int
	Status     string
	Type       string
	Currency   string
	FromDate   *time.Time
	ToDate     *time.Time
	FromAmount *float64
	ToAmount   *float64
}

type CurrencyRef struct {
	Code  string `json:"code"`
	Alias string `json:"alias"`
}

type Transaction struct {
	UUID            string      `json:"uuid"`
	Type            string      `json:"type"`
	ExternalAddress *string     `json:"external_address"`
	SourceCurrency  CurrencyRef `json:"source_currency"`
	DestCurrency    CurrencyRef `json:"dest_currency"`
	SourceAmount    float64     `json:"source_amount"`
	DestAmount      float64     `json:"dest_amount"`
	Status          string      `json:"status"`
	CreatedAt       string      `json:"created_at"`
	Comment         *string     `json:"comment"`
}

type ListData struct {
	Data        []Transaction `json:"data"`
	Pages       int           `json:"pages"`
	CurrentPage int           `json:"current_page"`
	Count       int           `json:"count"`
}

type ListResponse struct {
	Data ListData `json:"data"`
}

func (s *Service) List(ctx context.Context, req *ListRequest) (*transport.Response[ListResponse], error) {
	path := listPath
	if req != nil {
		q := url.Values{}
		if req.Page != nil {
			q.Set("page", strconv.Itoa(*req.Page))
		}
		if req.Limit != nil {
			q.Set("limit", strconv.Itoa(*req.Limit))
		}
		if req.Status != "" {
			q.Set("status", req.Status)
		}
		if req.Type != "" {
			q.Set("type", req.Type)
		}
		if req.Currency != "" {
			q.Set("currency", req.Currency)
		}
		if req.FromDate != nil {
			q.Set("from_date", req.FromDate.Format("2006-01-02"))
		}
		if req.ToDate != nil {
			q.Set("to_date", req.ToDate.Format("2006-01-02"))
		}
		if req.FromAmount != nil {
			q.Set("from_amount", strconv.FormatFloat(*req.FromAmount, 'f', -1, 64))
		}
		if req.ToAmount != nil {
			q.Set("to_amount", strconv.FormatFloat(*req.ToAmount, 'f', -1, 64))
		}
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &ListResponse{})
}
