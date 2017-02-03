package waitlivemsg

import (
	"fmt"
	"os"
	"strings"
	"time"

	"sync"

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
	Msg    string
	Stream *uilive.Writer

	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.Mutex
}

// Wait starts the waiting message
func (w *WaitLiveMsg) Wait() {
	var ticker = time.NewTicker(time.Second)
	w.tickerdMutex.Lock()
	w.tickerd = make(chan bool, 1)
	w.tickerdMutex.Unlock()
	w.start = time.Now()

	for {
		select {
		case t := <-ticker.C:
			var p = WarmupOn
			if t.Second()%2 == 0 {
				p = WarmupOff
			}

			var dots = strings.Repeat(".", t.Second()%3+1)

			fmt.Fprintf(w.Stream, "%c %v%s %ds\n",
				p, w.Msg, dots, w.Duration())

			if err := w.Stream.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "Error flushing startup ready message: %v\n", err)
			}
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
}

// Duration in seconds
func (w *WaitLiveMsg) Duration() int {
	return int(-w.start.Sub(time.Now()).Seconds())
}
