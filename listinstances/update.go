package listinstances

import (
	"github.com/wedeploy/cli/verbose"
	"golang.org/x/time/rate"
)

func (l *List) updateHandler() {
	var is, err = l.fetch()

	l.watchMutex.Lock()
	defer l.watchMutex.Unlock()

	switch {
	case err == nil:
	case isContextError(err):
		l.lastError = nil
		l.updated <- false
		return
	default:
		l.lastError = err
		l.retry++
		l.updated <- false
		return
	}

	l.Instances = is
	l.retry = 0
	l.updated <- true
}

func (l *List) update() {
	rate := rate.NewLimiter(rate.Every(l.PoolingInterval), 1)

	for {
		if err := rate.Wait(l.ctx); err != nil {
			verbose.Debug(err)
			return
		}

		l.updateHandler()
	}
}
