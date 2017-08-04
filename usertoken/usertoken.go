package usertoken

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/errwrap"
)

// JSONWebToken for the user
type JSONWebToken struct {
	Email string `json:"email"`
	UID   string `json:"sub"`
}

type jsonWebToken JSONWebToken

// Valid function for the JWT token
func (j jsonWebToken) Valid() error {
	return nil
}

// ParseJSONWebToken to retrieve an user info
func ParseJSONWebToken(accessToken string) (JSONWebToken, error) {
	var claims = jsonWebToken{}
	if _, err := jwt.ParseWithClaims(accessToken, &claims, keyFunc); err != nil {
		return JSONWebToken{}, getErrors(err)
	}

	return JSONWebToken(claims), nil
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	return []byte{}, nil
}

func getErrors(err error) error {
	// if only the bitmask for the 'signature invalid' is detected, ignore
	ev, ok := err.(*jwt.ValidationError)
	if ok && ev.Errors == jwt.ValidationErrorSignatureInvalid {
		return nil
	}

	return errwrap.Wrapf("Error parsing token: {{err}}", err)
}
