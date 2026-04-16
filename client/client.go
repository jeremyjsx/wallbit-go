package client

import (
	"errors"
	"strings"
)

var ErrMissingAPIKey = errors.New("wallbit client requires a non-empty api key")

type Client struct {
	apiKey string
	cfg    *Config
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

	return &Client{
		apiKey: apiKey,
		cfg:    cfg,
	}, nil
}

func (c *Client) Config() *Config {
	return c.cfg
}
