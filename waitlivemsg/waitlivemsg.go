package waitlivemsg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/henvic/uilive"
)

const (
	// WarmupOn symbol
	WarmupOn = '○'

	// WarmupOff symbol
	WarmupOff = '●'
)

// WaitLiveMsg is used for "waiting" live message
type WaitLiveMsg struct {
	msg          string
	stream       *uilive.Writer
	printMutex   sync.RWMutex
	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.Mutex
	startMutex   sync.RWMutex
	waitEnd      sync.WaitGroup
}

// SetMessage to display
func (w *WaitLiveMsg) SetMessage(msg string) {
	w.printMutex.Lock()
	w.msg = msg
	w.printMutex.Unlock()
}

// SetStream to output to
func (w *WaitLiveMsg) SetStream(ws *uilive.Writer) {
	w.printMutex.Lock()
	w.stream = ws
	w.printMutex.Unlock()
}

// Wait starts the waiting message
func (w *WaitLiveMsg) Wait() {
	w.waitEnd.Add(1)
	var ticker = time.NewTicker(time.Second)
	w.tickerdMutex.Lock()
	w.tickerd = make(chan bool, 1)
	w.tickerdMutex.Unlock()
	w.startMutex.Lock()
	w.start = time.Now()
	w.startMutex.Unlock()

	for {
		select {
		case t := <-ticker.C:
			w.printMutex.Lock()
			w.print(t)
			w.printMutex.Unlock()
		case <-w.tickerd:
			ticker.Stop()
			ticker = nil
			w.waitEnd.Done()
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

func (w *WaitLiveMsg) print(t time.Time) {
	var p = WarmupOn
	if t.Second()%2 == 0 {
		p = WarmupOff
	}

	var dots = strings.Repeat(".", t.Second()%3+1)

	fmt.Fprintf(w.stream, "%c %v%s %ds\n", p, w.msg, dots, w.Duration())
	_ = w.stream.Flush()
}
