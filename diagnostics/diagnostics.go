package diagnostics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/verbose"
)

// Diagnostics for the CLI and environment
type Diagnostics struct {
	Timeout     time.Duration
	Executables []*Executable
	Serial      bool
	ctx         context.Context
	cancel      context.CancelFunc
	queue       sync.WaitGroup
}

func (d *Diagnostics) exec(e *Executable) {
	var ctxCmd, ctxCmdCancel = context.WithTimeout(d.ctx, d.Timeout)
	defer func() {
		ctxCmdCancel()
		d.queue.Done()
	}()

	verbose.Debug(color.Format(color.FgHiYellow, "$ %s", e.Program()))
	var c = exec.CommandContext(ctxCmd, e.name, e.arg...)
	var b = &bytes.Buffer{}
	c.Stderr = b
	c.Stdout = b

	e.err = c.Run()

	var dm = color.Format(color.FgHiRed, `Terminated "%s"`, e.Program())

	if e.err != nil {
		dm += fmt.Sprintf(" (error: %v)", e.err)
	}

	verbose.Debug(dm)
	e.output = b.Bytes()

	if e.err != nil {
		e.output = append(e.output, []byte(fmt.Sprintln())...)
		e.output = append(e.output, []byte(fmt.Sprintln(e.err.Error()))...)
	}

	e.output = append(e.output, []byte(fmt.Sprintln())...)
}

func (d *Diagnostics) execAll() {
	d.queue.Add(len(d.Executables))

	// this should be just a wrapper for calling the list of executables
	// from Diagnostics and no other processing should be done here
	for _, e := range d.Executables {
		switch d.Serial {
		case true:
			d.exec(e)
		default:
			go d.exec(e)
		}
	}

	d.queue.Wait()
	d.cancel()
}

// Start diagnostics
func (d *Diagnostics) Start() (context.Context, context.CancelFunc) {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	go d.execAll()
	return d.ctx, d.cancel
}

// Report is a map of filename to content
type Report map[string][]byte

// Collect diagnostics
func (d *Diagnostics) Collect() Report {
	<-d.ctx.Done()

	var files = Report{}

	for _, e := range d.Executables {
		var appendTo = e.appendTo

		if appendTo == "" {
			appendTo = "log"
		}

		if _, ok := files[appendTo]; !ok {
			files[appendTo] = []byte{}
		}

		var what = []byte(fmt.Sprintln(color.Format(color.FgHiYellow, "$ %s", e.Program())))
		files[appendTo] = append(files[appendTo], what...)
		files[appendTo] = append(files[appendTo], e.output...)
	}

	return files
}

// Write report
func Write(w io.Writer, r Report) {
	for k, v := range r {
		fmt.Fprintf(os.Stderr,
			"%v%v%s",
			color.Format(color.Bold, color.BgHiBlue, " %v ", k),
			fmt.Sprintln(),
			v)
	}
}

// Executable is a command to be executed on the diagnostics
type Executable struct {
	appendTo string
	function func() error
	name     string
	arg      []string
	output   []byte
	err      error
}

// Program is the name of the program + arguments
func (e *Executable) Program() string {
	var program = e.name
	var args = strings.Join(e.arg, " ")

	if len(args) != 0 {
		program += " " + args
	}

	return program
}

func (e *Executable) Error() error {
	return e.err
}

// Entry for the diagnostics endpoint
type Entry struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Report   Report `json:"report"`
}

// Submit diagnostics
func Submit(ctx context.Context, entry Entry) error {
	// defaults.DiagnosticsEndpoint
	return errors.New("Not implemented")
}
