package assets

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jeremyjsx/wallbit-go/transport"
)

const listPath = "/api/public/v1/assets"

// ErrEmptySymbol is returned by [Service.Get] when symbol is empty or whitespace-only.
var ErrEmptySymbol = errors.New("assets: symbol is required")

type Service struct {
	sender transport.Sender
}

func NewService(sender transport.Sender) *Service {
	return &Service{sender: sender}
}

type ListRequest struct {
	Category string
	Search   string
	Page     *int
	Limit    *int
}

type Dividend struct {
	Amount      *float64 `json:"amount"`
	Yield       *float64 `json:"yield"`
	ExDate      *string  `json:"ex_date"`
	PaymentDate *string  `json:"payment_date"`
}

type Asset struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	AssetType     *string   `json:"asset_type"`
	Exchange      *string   `json:"exchange"`
	Sector        *string   `json:"sector"`
	MarketCapM    *string   `json:"market_cap_m"`
	Description   *string   `json:"description"`
	DescriptionES *string   `json:"description_es"`
	Country       *string   `json:"country"`
	CEO           *string   `json:"ceo"`
	Employees     *string   `json:"employees"`
	LogoURL       string    `json:"logo_url"`
	Dividend      *Dividend `json:"dividend"`
}

type ListResponse struct {
	Data        []Asset `json:"data"`
	Pages       int     `json:"pages"`
	CurrentPage int     `json:"current_page"`
	Count       int     `json:"count"`
}

type GetResponse struct {
	Data Asset `json:"data"`
}

func (s *Service) Get(ctx context.Context, symbol string) (*transport.Response[GetResponse], error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, ErrEmptySymbol
	}
	path := fmt.Sprintf("%s/%s", listPath, url.PathEscape(symbol))
	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &GetResponse{})
}

func (s *Service) List(ctx context.Context, req *ListRequest) (*transport.Response[ListResponse], error) {
	path := listPath
	if req != nil {
		q := url.Values{}
		if req.Category != "" {
			q.Set("category", req.Category)
		}
		if req.Search != "" {
			q.Set("search", req.Search)
		}
		if req.Page != nil {
			q.Set("page", strconv.Itoa(*req.Page))
		}
		if req.Limit != nil {
			q.Set("limit", strconv.Itoa(*req.Limit))
		}
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	return transport.SendJSON(ctx, s.sender, http.MethodGet, path, nil, &ListResponse{})
}
