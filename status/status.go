package status

import (
	"context"

	"github.com/wedeploy/cli/apihelper"
)

const (
	// Up status
	Up = "up"

	// Down status
	Down = "down"
)

type domains struct {
	Infrastructure string `json:"infrastructure"`
	Service        string `json:"service"`
}

// Status of the Backend
type Status struct {
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Domains domains `json:"domains"`
}

// IsUp is returned if the infrastructure is up
func IsUp(ctx context.Context) bool {
	var s, _ = get(ctx)
	return s.Status == Up
}

// UnsafeGet is a unsafe getter for the status
// See get() method below to understand why.
// It is exported here to use for domains.
func UnsafeGet(ctx context.Context) (s Status, err error) {
	return get(ctx)
}

func get(ctx context.Context) (s Status, err error) {
	// don't expose this method on the public API because
	// the current endpoint in unreliable, returning 503
	// with incompatible body responses by design
	// whenever status = "down", while the error structure
	// expects status to be a number and to have other info,
	// therefore I am not exposing it as it might cause problems later on
	var request = apihelper.URL(ctx, "/?options=verbose")
	if err = apihelper.Validate(request, request.Get()); err != nil {
		return s, err
	}

	err = apihelper.DecodeJSON(request, &s)
	return s, err
}
