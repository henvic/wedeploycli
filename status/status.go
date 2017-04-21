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

// Status of the Backend
type Status struct {
	Status string `json:"status"`
}

// IsUp is returned if the infrastructure is up
func IsUp(ctx context.Context) bool {
	var s, _ = get(ctx)
	return s.Status == Up
}

func get(ctx context.Context) (s Status, err error) {
	// don't expose this method on the public API because
	// the current endpoint in unreliable, returning 503
	// with incompatible body responses by design
	// whenever status = "down", while the error structure
	// expects status to be a number and to have other info,
	// therefore I am not exposing it as it might cause problems later on
	var request = apihelper.URL(ctx, "/")
	if err = apihelper.Validate(request, request.Get()); err != nil {
		return s, err
	}

	err = apihelper.DecodeJSON(request, &s)
	return s, err
}
