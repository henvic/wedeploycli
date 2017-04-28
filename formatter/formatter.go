package formatter

import (
	"io"
	"text/tabwriter"
)

// Human tells if the default formatting strategy should be human or machine friendly
var Human = false

// TabWriter for formatting text
type TabWriter struct {
	output    io.Writer
	tabwriter tabwriter.Writer
}

// NewTabWriter creates a TabWriter
func NewTabWriter(output io.Writer) *TabWriter {
	var t = &TabWriter{}
	t.Init(output)
	return t
}

// Init TabWriter
func (t *TabWriter) Init(output io.Writer) {
	t.output = output
	t.tabwriter.Init(output, 4, 0, 4, ' ', 0)
}

// Flush the TabWriter
func (t *TabWriter) Flush() error {
	return t.tabwriter.Flush()
}

// Write content
func (t *TabWriter) Write(buf []byte) (n int, err error) {
	if Human {
		return t.tabwriter.Write(buf)
	}

	return t.output.Write(buf)
}
