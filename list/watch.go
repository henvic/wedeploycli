package list

import (
	"context"
	"os"
	"time"

	"github.com/wedeploy/cli/errorhandling"
)

func (l *List) watchHandler() {
	<-l.updated

	defer func() {
		_ = l.w.Flush()
		_ = l.livew.Flush()
	}()

	if l.ctx.Err() == context.Canceled {
		return
	}

	l.watchMutex.RLock()
	var le = l.lastError
	var retry = l.retry
	l.watchMutex.RUnlock()

	if le != nil {
		l.Printf("%v #%d\n", errorhandling.Handle(le), retry)
		return
	}

	l.printProjects()
}

func (l *List) watchKiller(sigs chan os.Signal) {
	select {
	case <-sigs:
		l.stop()
	case <-l.ctx.Done():
		return
	}
}

func (l *List) watch() {
	l.w.Init(l.outStream)

	var ticker = time.NewTicker(l.PoolingInterval)

	for {
		select {
		case <-ticker.C:
			l.watchHandler()
		case <-l.ctx.Done():
			ticker.Stop()
			ticker = nil
			return
		}
	}
}
