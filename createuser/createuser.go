package createuser

import (
	"context"
	"errors"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/user"
)

// Try to create user
func Try(ctx context.Context) (err error) {
	_, err = user.Create(ctx, &user.User{
		Email:    "no-reply@wedeploy.com",
		Password: "cli-tool-password",
		Name:     "CLI Tool",
	})

	if err != nil {
		return errwrap.Wrapf("Failed to authenticate: {{err}}", err)
	}

	return createDeterministicToken()
}

func createDeterministicToken() error {
	// see https://github.com/wedeploy/cli/issues/307
	return errors.New("deterministic token required for local development not implemented")
}
