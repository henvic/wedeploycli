package list

import (
	"context"
	"fmt"
	"sync"

	"github.com/wedeploy/cli/verbose"
)

// RestartWatchList is a temporary watch
type RestartWatchList struct {
	Project           string
	Containers        []string
	IsStillRunning    func() bool
	list              *List
	projectHealthUID  string
	servicesHealthUID map[string]string
	healthMutex       sync.RWMutex
}

// Watch the RestartWatchList
func (rwl *RestartWatchList) Watch() {
	var queue sync.WaitGroup
	queue.Add(1)
	go func() {
		rwl.watchRoutine()
		queue.Done()
	}()
	queue.Wait()
}

// SetInitialProjectHealthUID sets the initial state for the HealthUID for the project
func (rwl *RestartWatchList) SetInitialProjectHealthUID(healthUID string) {
	rwl.healthMutex.Lock()
	rwl.projectHealthUID = healthUID
	rwl.healthMutex.Unlock()
}

// SetInitialContainersHealthUID sets the initial state for the HealthUID for the containers
func (rwl *RestartWatchList) SetInitialContainersHealthUID(healthUIDs map[string]string) {
	rwl.healthMutex.Lock()
	rwl.servicesHealthUID = healthUIDs
	rwl.healthMutex.Unlock()
}

func (rwl *RestartWatchList) watchRoutine() {
	var filter = Filter{
		Project:    rwl.Project,
		Containers: rwl.Containers,
	}

	rwl.list = New(filter)
	rwl.list.StopCondition = rwl.isDone
	rwl.list.Start()
}

func (rwl *RestartWatchList) isDone() bool {
	if rwl.IsStillRunning != nil && !rwl.IsStillRunning() {
		return false
	}

	if len(rwl.list.Projects) == 0 {
		verbose.Debug("Unexpected behavior: no projects found.")
		return false
	}

	var p = rwl.list.Projects[0]

	if p.Health != "up" || p.HealthUID == rwl.projectHealthUID {
		return false
	}

	var cs, ec = p.Services(context.Background())

	if ec != nil {
		fmt.Fprintf(rwl.list.outStream, "Can't check if containers are finished: %v\n", ec)
		return false
	}

	for _, c := range cs {
		current, ok := rwl.servicesHealthUID[c.ServiceID]
		if c.Health != "up" || (ok && current == c.HealthUID) {
			return false
		}
	}

	return true
}
