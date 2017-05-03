package waitlivemsg

import (
	"fmt"
	"sync"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
)

var spinners = []string{
	"⠋",
	"⠙",
	"⠹",
	"⠸",
	"⠼",
	"⠴",
	"⠦",
	"⠧",
	"⠇",
	"⠏",
}

// WaitLiveMsg is used for "waiting" live message
type WaitLiveMsg struct {
	msg          string
	stream       *uilive.Writer
	streamMutex  sync.RWMutex
	msgMutex     sync.RWMutex
	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.Mutex
	startMutex   sync.RWMutex
	waitEnd      sync.WaitGroup
}

// SetMessage to display
func (w *WaitLiveMsg) SetMessage(msg string) {
	w.msgMutex.Lock()
	w.msg = msg
	w.msgMutex.Unlock()
}

// SetStream to output to
func (w *WaitLiveMsg) SetStream(ws *uilive.Writer) {
	w.streamMutex.Lock()
	w.stream = ws
	w.streamMutex.Unlock()
}

// Wait starts the waiting message
func (w *WaitLiveMsg) Wait() {
	w.waitEnd.Add(1)
	w.tickerdMutex.Lock()
	w.tickerd = make(chan bool, 1)
	w.tickerdMutex.Unlock()
	w.startMutex.Lock()
	w.start = time.Now()
	w.startMutex.Unlock()

	w.waitLoop()

	w.msgMutex.RLock()
	w.print(w.msg, "✔")
	w.msgMutex.RUnlock()
	w.waitEnd.Done()
}

func (w *WaitLiveMsg) waitLoop() {
	var ticker = time.NewTicker(60 * time.Millisecond)
	var counter = 0
	for {
		select {
		case _ = <-ticker.C:
			w.msgMutex.RLock()
			w.print(w.msg, spinners[counter])
			counter = (counter + 1) % len(spinners)
			w.msgMutex.RUnlock()
		case <-w.tickerd:
			ticker.Stop()
			ticker = nil
			return
		}
	}
}

// Stop the waiting message
func (w *WaitLiveMsg) Stop() {
	w.tickerdMutex.Lock()
	w.tickerd <- true
	w.tickerdMutex.Unlock()
	w.waitEnd.Wait()
}

// StopWithMessage sets the last message and stops
func (w *WaitLiveMsg) StopWithMessage(msg string) {
	w.SetMessage(msg)
	w.Stop()
}

// ResetDuration to restart counter
func (w *WaitLiveMsg) ResetDuration() {
	w.startMutex.RLock()
	w.start = time.Now()
	w.startMutex.RUnlock()
}

// Duration in seconds
func (w *WaitLiveMsg) Duration() int {
	w.startMutex.RLock()
	var duration = int(-w.start.Sub(time.Now()).Seconds())
	w.startMutex.RUnlock()
	return duration
}

func (w *WaitLiveMsg) print(msg string, symbol string) {
	w.streamMutex.Lock()
	fmt.Fprintf(w.stream,
		"%v %v\n",
		color.Format(color.FgBlue, symbol),
		w.msg)
	_ = w.stream.Flush()
	w.streamMutex.Unlock()
}
