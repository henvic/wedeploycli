package deploy

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/wedeploy/cli/progress"
)

// writeCounter is a writer for writing to the progress bar
type writeCounter struct {
	Total    uint64
	Size     uint64
	progress *progress.Bar
}

// Write to the progress bar
func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	perc := uint64(progress.Total) * wc.Total / wc.Size

	wc.progress.Append = fmt.Sprintf(
		"%s/%s",
		humanize.Bytes(wc.Total),
		humanize.Bytes(wc.Size))

	wc.progress.Set(int(perc))

	return n, nil
}
