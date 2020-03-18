package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/henvic/wedeploycli/shell/internal/kubernetes"
	"github.com/henvic/wedeploycli/wesocket"
	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
)

// An ExitError reports an unsuccessful exit by a command.
type ExitError struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message,omitempty"`
	Details kubernetes.StatusDetails `json:"details,omitempty"`
}

// GetExitCode gets the exit code from the process.
func (e *ExitError) GetExitCode() (int, bool) {
	var details = e.Details

	for _, c := range details.Causes {
		if c.Type != "ExitCode" {
			continue
		}

		switch code, err := strconv.Atoi(c.Message); {
		case err != nil:
			return -1, false
		default:
			return code, true
		}
	}

	return 0, true
}

func (e *ExitError) Error() string {
	if code, ok := e.GetExitCode(); ok {
		return fmt.Sprintf("process terminated: %d", code)
	}

	var causes []string
	var details = e.Details

	for _, c := range details.Causes {
		causes = append(causes, fmt.Sprintf("%v: %v", c.Type, e.Message))
	}

	return fmt.Sprintf("cannot get exit code: %v\n%v", e.Message, strings.Join(causes, "\n"))
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
	t.Dialer = wesocket.Dialer()

	conn, err := gosocketio.ConnectContext(ctx, u, t)

	if err != nil {
		return err
	}

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
