package termsession

import (
	"context"
	"os"

	"github.com/henvic/socketio"
	"github.com/wedeploy/cli/verbose"
	"golang.org/x/crypto/ssh/terminal"
)

// TermSession to use on the shell.
type TermSession struct {
	ctx    context.Context
	cancel context.CancelFunc

	fd     int
	isTerm bool
	width  int
	height int
	state  *terminal.State

	watcher chan os.Signal

	socket *socketio.Namespace
}

// New terminal session.
func New(shell *socketio.Namespace) TermSession {
	return TermSession{
		fd:     int(os.Stdin.Fd()),
		socket: shell,
	}
}

// Start terminal session.
func (t *TermSession) Start(ctx context.Context, tty bool) (err error) {
	t.isTerm = tty

	if !t.isTerm {
		return nil
	}

	t.ctx, t.cancel = context.WithCancel(ctx)
	t.start()

	t.state, err = terminal.MakeRaw(t.fd)

	if err != nil {
		return err
	}

	if t.isTerm {
		go t.watchResize()
	}

	return nil
}

// Resize the terminal session if proportions changed.
func (t *TermSession) Resize() {
	if !t.isTerm {
		return
	}

	width, height, err := terminal.GetSize(t.fd)

	if err != nil {
		verbose.Debug("can't get resize dimensions:", err)
		return
	}

	if width == 0 || height == 0 || (width == t.width && height == t.height) {
		return
	}

	t.width = width
	t.height = height

	if t.socket == nil {
		verbose.Debug("Can't send resize terminal signal (socket not set)")
	}

	r := map[string]int{
		"width":  width,
		"height": height,
	}

	if err := t.socket.Emit("resize", &r); err != nil {
		verbose.Debug("error emitting resize message:", err)
	}
}

// Restore the terminal connected (close).
func (t *TermSession) Restore() error {
	if !t.isTerm {
		return nil
	}

	t.restore()

	err := terminal.Restore(t.fd, t.state)

	// not sure why this errno 0 keeps happening.
	// see https://github.com/golang/go/issues/24718
	if err == nil || err.Error() == "errno 0" {
		return nil
	}

	return err
}
