package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	ws "github.com/gorilla/websocket"
	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
)

// An ExitError reports an unsuccessful exit by a command.
type ExitError struct {
	PID      int `json:"pid"`
	ExitCode int `json:"exitCode"`
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("process terminated: %d", e.ExitCode)
}

// Params for the shell and its connection.
type Params struct {
	Host  string
	Token string

	ProjectID string
	ServiceID string

	Instance string

	AttachStdin bool
	TTY         bool
}

// Run shell command.
func Run(ctx context.Context, params Params, cmd string, args ...string) error {
	var process = &Process{
		Cmd:         cmd,
		Args:        args,
		TTY:         params.TTY,
		AttachStdin: params.AttachStdin,
	}

	var query = url.Values{}

	query.Add("accessToken", params.Token)
	query.Add("projectId", params.ProjectID)
	query.Add("serviceId", params.ServiceID)

	if params.Instance != "" {
		query.Add("containerId", params.Instance)
	}

	if !params.AttachStdin {
		query.Add("attachStdin", fmt.Sprint(params.AttachStdin))
	}

	if !params.TTY {
		query.Add("tty", fmt.Sprint(params.TTY))
	}

	if cmd != "" {
		query.Add("cmd", getCmdWithArgs(cmd, args))
	}

	var u = url.URL{
		Scheme:   "wss",
		Host:     params.Host,
		RawQuery: query.Encode(),
	}

	t := websocket.NewTransport()
	t.Dialer = websocketDialerConfig()
	conn, err := gosocketio.Connect(u, t)

	if err != nil {
		return err
	}

	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return process.Run(ctx, conn)
}

func websocketDialerConfig() ws.Dialer {
	// copy Proxy and TLSClientConfig options from the default WeDeploy client
	wedeployClient := wedeploy.Client()

	dialer := ws.Dialer{}
	transport := wedeployClient.HTTP().Transport

	if transport != nil {
		dt := transport.(*http.Transport)
		dialer.Proxy = dt.Proxy
		dialer.TLSClientConfig = dt.TLSClientConfig
	}

	return dialer
}

func getCmdWithArgs(cmd string, args []string) string {
	if len(cmd) != 0 {
		args = append([]string{cmd}, args...)
	}

	b, err := json.Marshal(args)

	if err != nil {
		panic(err)
	}

	return string(b)
}
