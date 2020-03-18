package list

import (
	"github.com/henvic/wedeploycli/verbose"
	"golang.org/x/time/rate"
)

func (l *List) updateHandler() {
	var ps, err = l.fetchProjects()

	l.watchMutex.Lock()
	defer l.watchMutex.Unlock()

	l.lastError = err

	if err != nil {
		l.retry++
		l.updated <- struct{}{}
		return
	}

	l.Projects = ps
	l.retry = 0
	l.updated <- struct{}{}
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
