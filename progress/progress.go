package progress

import (
	"strings"
	"time"

	"github.com/henvic/uiprogress"
)

// Bar holds the progress bar data
type Bar struct {
	// Name to prepend to the progress bar
	Name string
	// Prepend value to the progress bar after its name
	Prepend string
	// Append value to the progress bar
	Append string

	adapter *uiprogress.Bar
}

// Total of intervals
var Total = 100

// Width of standard progress bar
var Width = 40

var progressList = uiprogress.New()

// New progress bar instance
func New(name string) *Bar {
	var bar = &Bar{
		Name: name,
	}

	bar.setup()

	return bar
}

// Start progress bars
func Start() {
	progressList.Start()
}

// Stop progress bars
func Stop() {
	// allow enough time for the progress bar to print the final state
	time.Sleep(2 * progressList.RefreshInterval)
	progressList.Stop()
}

// Current position of the progress bar
func (b *Bar) Current() int {
	return b.adapter.Current()
}

// Flow repeats flowing the progress bar from begin to end continuously
func (b *Bar) Flow() {
	var current = b.Current()
	var next = current + 1

	if current >= Total-1 {
		next = 0
	}

	b.Set(next)
}

// Reset progress bar and reset its prepend and append messages
func (b *Bar) Reset(msgPrepend, msgAppend string) error {
	b.Prepend = msgPrepend
	b.Append = msgAppend
	return b.adapter.Set(0)
}

// Set progress bar position
func (b *Bar) Set(n int) error {
	// hack to show => even when complete
	if n == 100 {
		n = 99
	}

	return b.adapter.Set(n)
}

// Fail (give up)
func (b *Bar) Fail() {
	b.adapter.Head = byte('x')
}

func (b *Bar) setup() {
	b.adapter = progressList.AddBar(Total)
	b.adapter.Width = Width
	b.adapter.Empty = ' '

	b.adapter.PrependFunc(func(adapter *uiprogress.Bar) string {
		return strings.TrimSpace(b.Name + ": " + b.Prepend)
	})

	b.adapter.AppendFunc(func(adapter *uiprogress.Bar) string {
		return b.Append
	})
}
