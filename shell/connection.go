package shell

import (
	"errors"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/socketio"
	"github.com/wedeploy/cli/verbose"
)

func (p *Process) handleConnections() {
	if err := p.conn.On(socketio.OnDisconnect, func() {
		p.err <- errors.New("disconnected from gateway")
		p.ctxCancel()
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

	if err := p.conn.On(socketio.OnError, func(err error) {
		p.err <- errwrap.Wrapf("connection error: {{err}}", err)
		p.ctxCancel()
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

	if err := p.conn.On(socketio.OnConnection, func() {
		verbose.Debug("Connected to shell gateway.")
	}); err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

}
