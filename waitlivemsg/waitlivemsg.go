package waitlivemsg

import (
	"fmt"
	"sync"
	"time"

	"bytes"

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

const (
	tick  = "✔"
	cross = "✖"
)

// Message to print
type Message struct {
	text      string
	symbolEnd string
	end       bool
	counter   int
	noSymbol  bool
	mutex     sync.RWMutex
}

// NewMessage creates a Message with a given text
func NewMessage(text string) *Message {
	var m = Message{}
	m.SetText(text)
	return &m
}

// EmptyLine creates a Message with an empty line
// it has the side-effect of using no symbols as prefixes
func EmptyLine() *Message {
	return NewMessage("")
}

// NoSymbol hides the symbol for the given message
func (m *Message) NoSymbol() {
	m.mutex.Lock()
	m.noSymbol = true
	m.mutex.Unlock()
}

// SetText of message
func (m *Message) SetText(text string) {
	m.mutex.Lock()
	m.text = text
	m.mutex.Unlock()
}

func (m *Message) getSymbol() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.text == "" || m.noSymbol {
		return ""
	}

	var symbol = spinners[m.counter]
	m.counter = (m.counter + 1) % len(spinners)

	if !m.end {
		return symbol
	}

	if m.symbolEnd != "" {
		return m.symbolEnd
	}

	return GreenTickSymbol()
}

// SetSymbolEnd of message
func (m *Message) SetSymbolEnd(symbolEnd string) {
	m.mutex.Lock()
	m.symbolEnd = symbolEnd
	m.mutex.Unlock()
}

func (m *Message) isEnd() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.end
}

func (m *Message) getText() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.text
}

// End message
func (m *Message) End() {
	m.mutex.Lock()
	m.end = true
	m.mutex.Unlock()
}

// GreenTickSymbol for stop messages
func GreenTickSymbol() string {
	return color.Format(color.FgGreen, tick)
}

// RedCrossSymbol for stop messages
func RedCrossSymbol() string {
	return color.Format(color.FgRed, cross)
}

// WaitLiveMsg is used for "waiting" live message
type WaitLiveMsg struct {
	msgs         []*Message
	stream       *uilive.Writer
	streamMutex  sync.RWMutex
	msgsMutex    sync.RWMutex
	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.Mutex
	startMutex   sync.RWMutex
	waitEnd      sync.WaitGroup
}

// AddMessage to display
func (w *WaitLiveMsg) AddMessage(msg *Message) {
	w.msgsMutex.Lock()
	w.msgs = append(w.msgs, msg)
	w.msgsMutex.Unlock()
}

// SetMessage to display
func (w *WaitLiveMsg) SetMessage(msg *Message) {
	w.msgsMutex.Lock()
	w.msgs = []*Message{msg}
	w.msgsMutex.Unlock()
}

// SetMessages to display
func (w *WaitLiveMsg) SetMessages(msgs []*Message) {
	w.msgsMutex.Lock()
	w.msgs = msgs
	w.msgsMutex.Unlock()
}

// RemoveMessage from the messages slice
func (w *WaitLiveMsg) RemoveMessage(msg *Message) {
	w.msgsMutex.Lock()
	var newSlice = []*Message{}
	for _, m := range w.msgs {
		if m != msg {
			newSlice = append(newSlice, m)
		}
	}
	w.msgs = newSlice
	w.msgsMutex.Unlock()
}

// ResetMessages to display
func (w *WaitLiveMsg) ResetMessages() {
	w.msgsMutex.Lock()
	w.msgs = []*Message{}
	w.msgsMutex.Unlock()
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

	w.msgsMutex.RLock()
	w.print()
	w.msgsMutex.RUnlock()
	w.waitEnd.Done()
}

func (w *WaitLiveMsg) waitLoop() {
	var ticker = time.NewTicker(60 * time.Millisecond)
	for {
		select {
		case _ = <-ticker.C:
			w.msgsMutex.RLock()
			w.print()
			w.msgsMutex.RUnlock()
		case <-w.tickerd:
			ticker.Stop()
			ticker = nil
			return
		}
	}
}

// Stop the waiting message
func (w *WaitLiveMsg) Stop() {
	w.msgsMutex.RLock()
	for _, m := range w.msgs {
		if !m.isEnd() {
			m.End()
		}
	}
	w.msgsMutex.RUnlock()

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
func (w *WaitLiveMsg) Duration() time.Duration {
	w.startMutex.RLock()
	var duration = time.Now().Sub(w.start)
	w.startMutex.RUnlock()
	return duration
}

func (w *WaitLiveMsg) print() {
	w.streamMutex.Lock()

	var buf = bytes.Buffer{}

	for _, m := range w.msgs {
		var s = m.getSymbol()

		if len(s) != 0 {
			buf.WriteString(s)
			buf.WriteString(" ")
		}

		buf.WriteString(m.getText())
		buf.WriteString("\n")
	}

	fmt.Fprintf(w.stream, "%v", buf.String())
	_ = w.stream.Flush()
	w.streamMutex.Unlock()
}
