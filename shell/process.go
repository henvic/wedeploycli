package shell

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/socketio"
	"github.com/wedeploy/cli/color"
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

	if err := p.waitReady(); err != nil {
		return err
	}

	verbose.Debug("Connected to shell.")

	p.handleConnections()

	if err := p.Streams(); err != nil {
		return err
	}

	readyToStartExec := make(chan struct{}, 1)

	if err := p.shell.On("readyToStartExec", func() {
		readyToStartExec <- struct{}{}
	}); err != nil {
		return err
	}

	verbose.Debug("Waiting for 'readyToStartExec' signal")

	if err := p.waitReadyToStartExec(readyToStartExec); err != nil {
		return err
	}

	t := termsession.New(shell)

	if err := p.Fork(); err != nil {
		return err
	}

	if err := p.waitExecStarted(); err != nil {
		return err
	}

	defer stopTermSession(t)

	if err := t.Start(p.ctx, p.TTY); err != nil {
		return errwrap.Wrapf("can't initialize terminal: {{err}}", err)
	}

	err = <-p.err

	if err != nil {
		stopTermSession(t)

		// add a line break to separate connection errors from other messages
		_, _ = fmt.Fprintln(os.Stderr)
	}

	return err
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

	if err := p.shell.On("execStarted", func(es *execStarted) {
		p.printInfo(es)
		cerr <- nil
		p.execStarted <- struct{}{}
	}); err != nil {
		return err
	}

	return <-cerr
}

func (p *Process) waitReady() error {
	select {
	case err := <-p.err:
		return err
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.shell.Ready():
		verbose.Debug("Connection to the shell namespace is ready.")
		return nil
	}
}

func (p *Process) waitReadyToStartExec(readyToStartExec chan struct{}) error {
	select {
	case err := <-p.err:
		return err
	case <-readyToStartExec:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

func stopTermSession(t *termsession.TermSession) {
	err := t.Restore()

	if err != nil {
		err = errwrap.Wrapf("error trying to restore terminal: {{err}}", err)
	}

	verbose.Debug(err)
}

type execStarted struct {
	Instance string `json:"containerId,omitempty"`
}

func (p *Process) printInfo(es *execStarted) {
	trimmed := es.Instance

	if len(trimmed) >= 12 && !verbose.Enabled {
		trimmed = trimmed[:12]
	}

	var info = fmt.Sprintf("You are now accessing instance %s.\n",
		color.Format(color.FgMagenta, color.Bold, trimmed))

	if p.Cmd == "" {
		info += fmt.Sprintf("%s\n", color.Format(color.FgYellow,
			"Warning: don't use this shell to make changes on your services. Only changes inside volumes persist."))
	}

	if verbose.Enabled || p.Cmd != "" {
		verbose.Debug(info)
		return
	}

	_, _ = fmt.Fprintln(os.Stderr, info)
}
