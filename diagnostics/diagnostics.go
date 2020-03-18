package diagnostics

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	wedeploy "github.com/henvic/wedeploy-sdk-go"
	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/verbose"
	"github.com/henvic/wedeploycli/waitlivemsg"
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
			desc.StopText("Checked\t" + e.Description)
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
	var c = exec.CommandContext(ctxCmd, name, args...) // #nosec
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

// Report of the system diagnostics.
type Report []byte

// Collect diagnostics
func (d *Diagnostics) Collect() Report {
	<-d.ctx.Done()

	var r Report

	for _, e := range d.Executables {
		var what = []byte(fmt.Sprintf("$ %s\n", e.Command))
		r = append(r, what...)
		r = append(r, e.output...)

		if r[len(r)-1] != '\n' {
			r = append(r, byte('\n'))
		}
	}

	return r
}

// Executable is a command to be executed on the diagnostics
type Executable struct {
	Description string
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
	ID       string `json:"id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Report   string `json:"report"`
}

// Submit diagnostics
func Submit(ctx context.Context, entry Entry) (err error) {
	var req = wedeploy.URL(defaults.DiagnosticsEndpoint)
	req.SetContext(ctx)

	err = apihelper.SetBody(req, submitPost{
		ID:       entry.ID,
		Username: entry.Username,
		Time:     time.Now().Format(time.RubyDate),
		Report:   string(entry.Report),
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
