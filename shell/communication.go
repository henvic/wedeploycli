package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/verbose"
)

// Fork the process
func (p *Process) Fork() error {
	verbose.Debug("Forking process!")

	// for some reason the server seems to require tty to be passed here as well
	startExecOptions := map[string]bool{}

	return p.shell.Emit("startExec", startExecOptions)
}

// Streams (stdin, stderr, stdout, end channel) from/to UNIX socket/websocket
func (p *Process) Streams() error {
	var streams = [](func() error){
		p.PipeStderr,
		p.MaybePipeStdin,
		p.PipeStdout,
		p.WatchEnd,
		p.WatchErrors,
	}

	for _, s := range streams {
		if err := s(); err != nil {
			return err
		}
	}

	return nil
}

// MaybePipeStdin from websocket to UNIX socket
func (p *Process) MaybePipeStdin() error {
	if !p.AttachStdin {
		return nil
	}

	return p.PipeStdin()
}

// PipeStdin from websocket to UNIX socket
func (p *Process) PipeStdin() (err error) {
	go p.pipeStdinGoroutine()
	return nil
}

func (p *Process) pipeStdinGoroutine() {
	var inStream io.ReadCloser = os.Stdin
	reader := bufio.NewReader(inStream)
	defer func() {
		_ = inStream.Close()
	}()

	select {
	case <-p.ctx.Done():
		return
	case <-p.execStarted:
	}
	// probably want to listen for SIGTERM

readStdin:
	if p.ctx.Err() != nil {
		return
	}

	b, _, err := reader.ReadRune()

	if err == io.EOF {
		verbose.Debug("Closing stdin: reading rune returned io.EOF")

		if err := p.shell.Emit("stdinDone", map[string]string{}); err != nil {
			verbose.Debug("error sending stdinEOF signal:", err)
		}
		return
	}

	if err != nil {
		p.err <- err
		p.ctxCancel()
		return
	}

	var bg = []byte(string(b))
	bfed := reader.Buffered()

	if bfed != 0 {
		if err := reader.UnreadRune(); err != nil {
			p.err <- errwrap.Wrapf("stdin unread rune issue: {{err}}", err)
			p.ctxCancel()
			return
		}

		// peeking the whole stdin at once, but maybe choosing a chunk size to slice it is wiser
		bg, err = reader.Peek(bfed)

		if err != nil {
			p.err <- errwrap.Wrapf("stdin peeking issue: {{err}}", err)
			p.ctxCancel()
			return
		}

		if _, err := reader.Discard(len(bg)); err != nil {
			p.err <- errwrap.Wrapf("stdin discarding issue: {{err}}", err)
			p.ctxCancel()
			return
		}
	}

	if err := p.shell.Emit("stdin", string(bg)); err != nil {
		p.err <- errwrap.Wrapf("stdin pipe broken: {{err}}", err)
		p.ctxCancel()
		return
	}

	// a sleep() throttle call might go here
	goto readStdin
}

// PipeStdout from UNIX socket to websocket
func (p *Process) PipeStdout() error {
	return p.shell.On("stdout", func(content string) {
		fmt.Print(content)
	})
}

// PipeStderr from UNIX socket to websocket
func (p *Process) PipeStderr() error {
	return p.shell.On("stderr", func(content string) {
		fmt.Fprint(os.Stderr, content)
	})
}

// WatchEnd of process
func (p *Process) WatchEnd() error {
	return p.shell.On("execExit", func(e *ExitError) {
		verbose.Debug("Process", e.PID, "exited.")

		if e.ExitCode == 0 {
			p.err <- nil
			p.ctxCancel()
			return
		}

		p.err <- e
		p.ctxCancel()
	})
}

// WatchErrors of the socketio module
func (p *Process) WatchErrors() error {
	return p.shell.On("error", func(err error) {
		verbose.Debug("shell socket error:", err)
	})
}
