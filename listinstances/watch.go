package listinstances

import (
	"time"

	"github.com/henvic/ctxsignal"
	"github.com/henvic/wedeploycli/errorhandler"
)

func (l *List) watchHandler() {
	select {
	case <-l.ctx.Done():
		return
	case <-l.updated:
	}

	defer func() {
		_ = l.w.Flush()
		_ = l.livew.Flush()
	}()

	if _, err := ctxsignal.Closed(l.ctx); err == nil {
		return
	}

	l.watchMutex.RLock()
	var le = l.lastError
	var retry = l.retry
	l.watchMutex.RUnlock()

	if le != nil {
		if l.once {
			return
		}

		l.Printf("%v #%d\n", errorhandler.Handle(le), retry)
		return
	}

	l.printInstances()
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
