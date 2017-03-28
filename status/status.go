package status

import (
	"context"

	"github.com/wedeploy/cli/apihelper"
)

// Status of the Backend
type Status struct {
	Version string `json:"version"`
}

// Get current status
func Get(ctx context.Context) (Status, error) {
	var status Status
	var err = apihelper.AuthGet(ctx, "/", &status)
	return status, err
}
