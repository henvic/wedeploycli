package shell

import (
	"errors"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/gosocketio"
)

func (p *Process) handleConnections() {
	if err := p.conn.On(gosocketio.OnDisconnect, func() {
		p.err <- errors.New("disconnected from gateway")
		p.ctxCancel()
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

	if err := p.conn.On(gosocketio.OnError, func(err error) {
		p.err <- errwrap.Wrapf("connection error: {{err}}", err)
		p.ctxCancel()
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

	if err := p.conn.On(gosocketio.OnConnection, func() {
		verbose.Debug("Connected to shell gateway.")
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

}
