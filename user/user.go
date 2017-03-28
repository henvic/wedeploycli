package user

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
)

// User on the infrastructure
type User struct {
	ID       string `json:"id,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Create user
func Create(ctx context.Context, u *User) (*User, error) {
	var request = apihelper.URL(ctx, "/user/create")
	apihelper.Auth(request)
	var b, err = json.Marshal(u)

	if err != nil {
		return nil, errwrap.Wrapf("Can not create User JSON request: {{err}}", err)
	}

	request.Body(bytes.NewBuffer(b))

	if err = apihelper.Validate(request, request.Post()); err != nil {
		return nil, err
	}

	var ur = &User{}
	err = apihelper.DecodeJSON(request, &ur)
	return ur, err
}
