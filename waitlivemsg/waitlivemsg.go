package waitlivemsg

import (
	"fmt"
	"sync"
	"time"

	"bytes"

	"github.com/henvic/uilive"
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

// Message to print
type Message struct {
	text    string
	counter int
	mutex   sync.RWMutex
}

// NewMessage creates a Message with a given text
func NewMessage(text string) *Message {
	var m = Message{}
	m.PlayText(text)
	return &m
}

// EmptyLine creates a Message with an empty line
// it has the side-effect of using no symbols as prefixes
func EmptyLine() *Message {
	return NewMessage("")
}

// PlayText as a live message
func (m *Message) PlayText(text string) {
	m.mutex.Lock()
	m.text = text
	if m.counter == -1 {
		m.counter = 0
	}
	m.mutex.Unlock()
}

// StopText of the live messages
func (m *Message) StopText(text string) {
	m.mutex.Lock()
	m.text = text
	m.counter = -1
	m.mutex.Unlock()
}

// GetText of the message
func (m *Message) GetText() string {
	m.mutex.RLock()
	var text = m.text
	m.mutex.RUnlock()
	return text
}

func (m *Message) getSymbol() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.counter == -1 || m.text == "" {
		return ""
	}

	var symbol = spinners[m.counter]
	m.counter = (m.counter + 1) % len(spinners)

	return symbol
}

func (m *Message) getText() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.text
}

// End of live message
func (m *Message) End() {
	m.mutex.Lock()
	m.counter = -1
	m.mutex.Unlock()
}

// WaitLiveMsg is used for "waiting" live message
type WaitLiveMsg struct {
	msgs         []*Message
	stream       *uilive.Writer
	streamMutex  sync.RWMutex
	msgsMutex    sync.RWMutex
	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.RWMutex
	startMutex   sync.RWMutex
	waitEnd      sync.WaitGroup
}

// New creates a WaitLiveMsg
func New(ws *uilive.Writer) *WaitLiveMsg {
	if ws == nil {
		ws = uilive.New()
	}

	return &WaitLiveMsg{
		stream: ws,
	}
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
		w.tickerdMutex.RLock()
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
		w.tickerdMutex.RUnlock()
	}
}

// Stop the waiting message
func (w *WaitLiveMsg) Stop() {
	w.msgsMutex.RLock()
	for _, m := range w.msgs {
		m.mutex.RLock()
		c := m.counter
		m.mutex.RUnlock()
		if c != -1 {
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
