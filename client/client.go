package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/jeremyjsx/wallbit-go/internal/errorsx"
	"github.com/jeremyjsx/wallbit-go/services/apikey"
	"github.com/jeremyjsx/wallbit-go/services/balance"
	"github.com/jeremyjsx/wallbit-go/services/fees"
	"github.com/jeremyjsx/wallbit-go/services/trades"
	"github.com/jeremyjsx/wallbit-go/services/transactions"
)

var ErrMissingAPIKey = errors.New("wallbit client requires a non-empty api key")

type Client struct {
	apiKey string
	cfg    *Config

	Balance *balance.Service

	Transactions *transactions.Service

	APIKey *apikey.Service

	Trades *trades.Service

	Fees *fees.Service
}

func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrMissingAPIKey
	}

	cfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	c := &Client{
		apiKey: apiKey,
		cfg:    cfg,
	}
	c.Balance = balance.NewService(c)
	c.Transactions = transactions.NewService(c)
	c.APIKey = apikey.NewService(c)
	c.Trades = trades.NewService(c)
	c.Fees = fees.NewService(c)

	return c, nil
}

func (c *Client) Config() *Config {
	return c.cfg
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	endpoint, err := c.cfg.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) send(ctx context.Context, method string, path string, body io.Reader, dest any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	return c.do(req, dest)
}

func (c *Client) Send(ctx context.Context, method string, path string, body io.Reader, dest any) error {
	return c.send(ctx, method, path, body, dest)
}

func (c *Client) do(req *http.Request, dest any) error {
	res, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errorsx.FromHTTP(res.StatusCode, res.Header.Get("X-Request-ID"), body)
	}

	if dest == nil || len(body) == 0 || res.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return err
	}

	return nil
}
