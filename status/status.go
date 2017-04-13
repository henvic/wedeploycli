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

// Get current status
func Get(ctx context.Context) (s Status, err error) {
	var request = apihelper.URL(ctx, "/")
	if err = apihelper.Validate(request, request.Get()); err != nil {
		return s, err
	}

	err = apihelper.DecodeJSON(request, &s)
	return s, err
}
