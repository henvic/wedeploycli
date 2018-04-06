// +build !windows

package termsession

import (
	"os"
	"os/signal"
	"syscall"
)

func (t *TermSession) start() {
	t.watcher = make(chan os.Signal, 1)
	signal.Notify(t.watcher, syscall.SIGWINCH)
}

func (t *TermSession) watchResize() {
	t.Resize()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.watcher:
			t.Resize()
		}
	}
}

func (t *TermSession) restore() {
	signal.Reset(syscall.SIGWINCH)
}
