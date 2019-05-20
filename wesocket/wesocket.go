// Package wesocket handles websocket connections to the cloud infrastructure.
package wesocket

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/wedeploy/gosocketio"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
)

// Dialer for websocket connection with default client settings.
func Dialer() websocket.Dialer {
	client := wedeploy.Client()
	dialer := websocket.Dialer{}
	transport := client.HTTP().Transport

	if transport != nil {
		dt := transport.(*http.Transport)
		dialer.Proxy = dt.Proxy
		dialer.TLSClientConfig = dt.TLSClientConfig
	}

	return dialer
}

// Authenticate resolves the authentication for the socket.io connection.
func Authenticate(conn *gosocketio.Namespace) error {
	var cerr = make(chan error, 1)

	if err := conn.On("authentication", func(a authMap) {
		if !a.Success {
			cerr <- errors.New("server authentication failure")
			return
		}

		cerr <- nil
	}); err != nil {
		return err
	}

	return <-cerr
}

type authMap struct {
	Success bool
}
