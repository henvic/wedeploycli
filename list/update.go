package list

import "time"

func (l *List) updateHandler() {
	var ps, err = l.fetchProjects()

	switch {
	case err == nil:
	case isContextError(err):
		l.lastError = nil
		l.updated <- false
		return
	default:
		l.watchMutex.Lock()
		l.lastError = err
		l.retry++
		l.updated <- false
		l.watchMutex.Unlock()
		return
	}

	l.watchMutex.Lock()
	l.Projects = ps
	l.retry = 0
	l.updated <- true
	l.watchMutex.Unlock()
}

func (l *List) update() {
	for {
		select {
		default:
			l.updateHandler()
			time.Sleep(time.Second)
		case <-l.ctx.Done():
			return
		}
	}
}
