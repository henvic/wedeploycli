package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/henvic/socketio"
	"github.com/henvic/socketio/websocket"
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

	TTY     bool
	Detach  bool
	NoStdin bool
}

// Run shell command.
func Run(ctx context.Context, params Params, cmd string, args ...string) error {
	var process = &Process{
		Cmd:     cmd,
		Args:    args,
		TTY:     params.TTY,
		NoStdin: params.NoStdin,
	}

	var query = url.Values{}

	query.Add("accessToken", params.Token)
	query.Add("projectId", params.ProjectID)
	query.Add("serviceId", params.ServiceID)
	query.Add("tty", fmt.Sprint(params.TTY))
	query.Add("detach", fmt.Sprint(params.Detach))

	if cmd != "" {
		query.Add("cmd", getCmdWithArgs(cmd, args))
	}

	var u = url.URL{
		Scheme:   "wss",
		Host:     params.Host,
		RawQuery: query.Encode(),
	}

	conn, err := socketio.Dial(u, websocket.NewTransport())

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
