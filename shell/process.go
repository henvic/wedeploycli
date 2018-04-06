package shell

import (
	"context"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/socketio"
	"github.com/wedeploy/cli/shell/internal/termsession"
	"github.com/wedeploy/cli/verbose"
)

// Process to run
type Process struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	Cmd  string
	Args []string

	TTY     bool
	NoStdin bool

	PID      int
	ExitCode int

	conn  *socketio.Client
	shell *socketio.Namespace

	err chan error
}

// Run connection
func (p *Process) Run(ctx context.Context, conn *socketio.Client) (err error) {
	p.ctx, p.ctxCancel = context.WithCancel(ctx)

	shell, err := conn.Of("/subscribe/project/service/container")

	if err != nil {
		return err
	}

	verbose.Debug("Connected to shell.")

	p.conn = conn
	p.shell = shell
	p.err = make(chan error, 1)

	p.handleConnections()

	if err := p.Streams(); err != nil {
		return err
	}

	ts := termsession.New(shell)

	defer func() {
		e := ts.Restore()

		if e != nil {
			e = errwrap.Wrapf("error trying to restore terminal: {{err}}", e)
		}

		if err != nil {
			verbose.Debug(e)
			return
		}

		err = e
	}()

	if err := ts.Start(p.ctx, p.TTY); err != nil {
		return errwrap.Wrapf("can't initialize terminal: {{err}}", err)
	}

	if err := p.Fork(); err != nil {
		return err
	}

	return <-p.err
}
