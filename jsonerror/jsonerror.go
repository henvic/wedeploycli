package jsonerror

import (
	"encoding/json"

	"github.com/hashicorp/errwrap"
)

// FriendlyUnmarshal simplifies a JSON Unmarshal error message.
func FriendlyUnmarshal(err error) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*json.UnmarshalTypeError)

	if !ok {
		return err
	}

	if e.Struct != "" || e.Field != "" {
		return errwrap.Wrapf("json: cannot decode "+e.Value+" as "+e.Struct+" \""+e.Field+"\" "+e.Type.String()+" field", err)
	}

	return errwrap.Wrapf("json: cannot decode "+e.Value+" as "+e.Type.String(), err)
}
