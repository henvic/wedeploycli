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
	var c = m.counter
	m.mutex.RUnlock()

	if c == -1 || m.text == "" {
		return ""
	}

	var symbol = spinners[c]
	m.mutex.Lock()
	m.counter = (c + 1) % len(spinners)
	m.mutex.Unlock()

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
	msgsMutex    sync.RWMutex
	start        time.Time
	tickerd      chan bool
	tickerdMutex sync.RWMutex
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

// GetMessages displayed
func (w *WaitLiveMsg) GetMessages() []*Message {
	w.msgsMutex.RLock()
	defer w.msgsMutex.RUnlock()
	return w.msgs
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
	w.msgsMutex.Lock()
	w.stream = ws
	w.msgsMutex.Unlock()
}

// Wait starts the waiting message
func (w *WaitLiveMsg) Wait() {
	w.waitEnd.Add(1)
	w.tickerdMutex.Lock()
	w.tickerd = make(chan bool, 1)
	w.start = time.Now()
	w.tickerdMutex.Unlock()

	w.waitLoop()

	w.print()
	w.waitEnd.Done()
}

func (w *WaitLiveMsg) waitLoop() {
	var ticker = time.NewTicker(60 * time.Millisecond)
	for {
		w.tickerdMutex.RLock()
		select {
		case <-ticker.C:
			w.print()
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

	if w.tickerd != nil {
		w.tickerd <- true
	}

	w.tickerdMutex.Unlock()
	w.waitEnd.Wait()
}

// ResetDuration to restart counter
func (w *WaitLiveMsg) ResetDuration() {
	w.tickerdMutex.RLock()
	w.start = time.Now()
	w.tickerdMutex.RUnlock()
}

// Duration in seconds
func (w *WaitLiveMsg) Duration() time.Duration {
	w.tickerdMutex.RLock()
	var duration = time.Since(w.start)
	w.tickerdMutex.RUnlock()
	return duration
}

func (w *WaitLiveMsg) print() {
	var buf = bytes.Buffer{}

	w.msgsMutex.RLock()
	var msgs = w.msgs
	w.msgsMutex.RUnlock()

	for _, m := range msgs {
		var txt = m.getText()

		if len(txt) == 0 {
			continue
		}

		var s = m.getSymbol()

		if len(s) != 0 {
			buf.WriteString(s)
			buf.WriteString(" ")
		}

		buf.WriteString(txt)
		buf.WriteString("\n")
	}

	w.msgsMutex.Lock()
	defer w.msgsMutex.Unlock()
	_, _ = fmt.Fprintf(w.stream, "%v", buf.String())
	_ = w.stream.Flush()
}
