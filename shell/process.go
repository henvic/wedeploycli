package shell

import (
	"context"
	"errors"

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

	TTY         bool
	AttachStdin bool

	PID      int
	ExitCode int

	conn  *socketio.Client
	shell *socketio.Namespace

	execStarted chan struct{}

	err chan error
}

// Run connection
func (p *Process) Run(ctx context.Context, conn *socketio.Client) (err error) {
	p.ctx, p.ctxCancel = context.WithCancel(ctx)
	defer p.ctxCancel()

	shell, err := conn.Of("/subscribe/project/service/container")

	if err != nil {
		return err
	}

	p.conn = conn
	p.shell = shell
	p.execStarted = make(chan struct{}, 1)
	p.err = make(chan error, 1)

	if err := p.authenticate(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.shell.Ready():
		verbose.Debug("Connection to the shell namespace is ready.")
		break
	}

	verbose.Debug("Connected to shell.")

	p.handleConnections()

	if err := p.Streams(); err != nil {
		return err
	}

	canFork := make(chan struct{}, 1)

	if err := p.shell.On("readyToStartExec", func() {
		canFork <- struct{}{}
	}); err != nil {
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

	verbose.Debug("Waiting for 'readyToStartExec' signal")

	select {
	case <-canFork:
	case <-p.ctx.Done():
		return p.ctx.Err()
	}

	if err := p.Fork(); err != nil {
		return err
	}

	if err := p.waitExecStarted(); err != nil {
		return err
	}

	return <-p.err
}

type authMap struct {
	Success bool
}

func (p *Process) authenticate() error {
	var cerr = make(chan error, 1)

	if err := p.shell.On("authentication", func(a authMap) {
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

func (p *Process) waitExecStarted() error {
	var cerr = make(chan error, 1)

	if err := p.shell.On("execStarted", func() {
		cerr <- nil
		p.execStarted <- struct{}{}
	}); err != nil {
		return err
	}

	return <-cerr
}
