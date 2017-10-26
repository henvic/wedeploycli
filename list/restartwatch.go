package list

import (
	"context"
	"fmt"
	"sync"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

// RestartWatchList is a temporary watch
type RestartWatchList struct {
	Project           string
	Services          []string
	IsStillRunning    func() bool
	list              *List
	projectHealthUID  string
	servicesHealthUID map[string]string
	healthMutex       sync.RWMutex
	wectx             config.Context
}

// Watch the RestartWatchList
func (rwl *RestartWatchList) Watch(wectx config.Context) {
	rwl.wectx = wectx
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

// SetInitialServicesHealthUID sets the initial state for the HealthUID for the services
func (rwl *RestartWatchList) SetInitialServicesHealthUID(healthUIDs map[string]string) {
	rwl.healthMutex.Lock()
	rwl.servicesHealthUID = healthUIDs
	rwl.healthMutex.Unlock()
}

func (rwl *RestartWatchList) watchRoutine() {
	var filter = Filter{
		Project:  rwl.Project,
		Services: rwl.Services,
	}

	rwl.list = New(filter)
	rwl.list.StopCondition = rwl.isDone
	rwl.list.Start(rwl.wectx)
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

	if p.Health == "empty" {
		return true
	}

	if p.Health != "up" || p.HealthUID == rwl.projectHealthUID {
		return false
	}

	servicesClient := services.New(rwl.list.wectx)

	var cs, ec = p.Services(context.Background(), servicesClient)

	if ec != nil {
		fmt.Fprintf(rwl.list.outStream, "Can't check if services are finished: %v\n", ec)
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
