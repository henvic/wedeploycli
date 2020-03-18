package status

import (
	"context"

	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/config"
)

const (
	// Up status
	Up = "up"

	// Down status
	Down = "down"
)

// Client for the services
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		apihelper.New(wectx),
	}
}

// Domains of the infrastructure
type Domains struct {
	Infrastructure string `json:"infrastructure"`
	Service        string `json:"service"`
}

// Status of the Backend
type Status struct {
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Domains Domains `json:"domains"`
}

// IsUp is returned if the infrastructure is up
func (c *Client) IsUp(ctx context.Context) bool {
	var s, _ = c.get(ctx)
	return s.Status == Up
}

// UnsafeGet is a unsafe getter for the status
// See get() method below to understand why.
// It is exported here to use for domains.
func (c *Client) UnsafeGet(ctx context.Context) (s Status, err error) {
	return c.get(ctx)
}

func (c *Client) get(ctx context.Context) (s Status, err error) {
	// don't expose this method on the public API because
	// the current endpoint in unreliable, returning 503
	// with incompatible body responses by design
	// whenever status = "down", while the error structure
	// expects status to be a number and to have other info,
	// therefore I am not exposing it as it might cause problems later on
	var request = c.Client.URL(ctx, "/")
	if err = apihelper.Validate(request, request.Get()); err != nil {
		return s, err
	}

	err = apihelper.DecodeJSON(request, &s)
	return s, err
}
