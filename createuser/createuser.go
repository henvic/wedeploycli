package createuser

import (
	"context"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/user"
)

func Try(ctx context.Context) (err error) {
	_, err = user.Create(ctx, &user.User{
		Email:    "no-reply@wedeploy.com",
		Password: "cli-tool-password",
		Name:     "CLI Tool",
	})

	if err != nil {
		return errwrap.Wrapf("Failed to authenticate: {{err}}", err)
	}

	return err
}
