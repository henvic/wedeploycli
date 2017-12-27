package diagnostics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
)

// Diagnostics for the CLI and environment
type Diagnostics struct {
	Timeout     time.Duration
	Executables []*Executable
	Serial      bool
	ctx         context.Context
	cancel      context.CancelFunc
	queue       sync.WaitGroup
	wlm         waitlivemsg.WaitLiveMsg
}

func (d *Diagnostics) exec(e *Executable) {
	var ctxCmd, ctxCmdCancel = context.WithTimeout(d.ctx, d.Timeout)
	defer func() {
		ctxCmdCancel()
	}()

	var desc = waitlivemsg.NewMessage("Checking\t" + e.Description)
	defer func() {
		if e.IgnoreError || e.err == nil {
			desc.StopText(fancy.Success("Checked\t") + e.Description)
		} else {
			desc.StopText(fancy.Error("Error\t") + e.Description)
		}
	}()

	// only add message if it is not empty
	if e.Description != "" {
		d.wlm.AddMessage(desc)
	}

	verbose.Debug(color.Format(color.FgHiYellow, "$ %s", e.Command))
	var name, args = getRunCommand(e.Command)
	var c = exec.CommandContext(ctxCmd, name, args...)
	var b = &bytes.Buffer{}
	c.Stderr = b
	c.Stdout = b

	e.err = c.Run()

	var dm = color.Format(color.FgHiRed, `Terminated "%s"`, e.Command)

	if e.err != nil {
		if e.IgnoreError {
			dm += fmt.Sprintf(" (error: %v; probably safe to ignore due to command not available on system)", e.err)
		} else {
			dm += fmt.Sprintf(" (error: %v)", e.err)
		}
	}

	verbose.Debug(dm)
	e.output = b.Bytes()

	if e.err != nil {
		e.output = append(e.output, []byte(fmt.Sprintf("\n%v\n", e.err))...)
	}

	e.output = append(e.output, []byte(fmt.Sprintln())...)
}

func (d *Diagnostics) execAll() {
	d.wlm = *waitlivemsg.New(nil)
	go d.wlm.Wait()
	defer d.wlm.Stop()

	d.queue.Add(len(d.Executables))

	// we don't want to fire dozen of commands at once
	var concurrencyLimit = 10

	if d.Serial {
		concurrencyLimit = 1
	}

	// idea from the net package
	// see http://jmoiron.net/blog/limiting-concurrency-in-go/
	sem := make(chan struct{}, concurrencyLimit)

	for _, e := range d.Executables {
		sem <- struct{}{}
		go func(re *Executable) {
			defer func() {
				<-sem
			}()
			d.exec(re)
			d.queue.Done()
		}(e)
	}

	d.queue.Wait()
	verbose.Debug("Finished executing all diagnostic commands")
	d.cancel()
}

// Run diagnostics
func (d *Diagnostics) Run(ctx context.Context) {
	d.ctx, d.cancel = context.WithCancel(ctx)
	d.execAll()
}

// Report is a map of filename to content
type Report map[string][]byte

// Len returns the number of bytes of a report
func (r *Report) Len() int {
	var l = 0

	for _, rb := range *r {
		l += len(rb)
	}

	return l
}

// String created from the report
func (r *Report) String() map[string]string {
	var m = map[string]string{}

	for k, rb := range *r {
		m[k] = string(rb)
	}

	return m
}

// Collect diagnostics
func (d *Diagnostics) Collect() Report {
	<-d.ctx.Done()

	var files = Report{}

	for _, e := range d.Executables {
		if e.IgnoreError {
			break
		}

		var appendTo = e.LogFile

		if appendTo == "" {
			appendTo = "log"
		}

		if _, ok := files[appendTo]; !ok {
			files[appendTo] = []byte{}
		}

		var what = []byte(fmt.Sprintln(color.Format(color.FgHiYellow, "$ %s", e.Command)))
		files[appendTo] = append(files[appendTo], what...)
		files[appendTo] = append(files[appendTo], e.output...)
	}

	return files
}

// Write report
func Write(w io.Writer, r Report) {
	for k, v := range r {
		fmt.Fprintf(os.Stderr,
			"%v\n%s",
			color.Format(color.Bold, color.BgHiBlue, " %v ", k),
			v)
	}
}

// Executable is a command to be executed on the diagnostics
type Executable struct {
	Description string
	LogFile     string
	Command     string
	// IgnoreError if command exit code is != 0 (don't log)
	IgnoreError bool
	Required    bool
	output      []byte
	err         error
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

type submitPost struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Report   map[string]string `json:"report"`
}

// Submit diagnostics
func Submit(ctx context.Context, entry Entry) (err error) {
	var req = wedeploy.URL(defaults.DiagnosticsEndpoint)
	req.SetContext(ctx)

	err = apihelper.SetBody(req, submitPost{
		ID:       entry.ID,
		Username: entry.Username,
		Report:   entry.Report.String(),
	})

	if err != nil {
		return err
	}

	err = req.Post()
	return apihelper.Validate(req, err)
}

func checkBashExists() bool {
	_, err := exec.LookPath("bash")
	return err == nil
}
