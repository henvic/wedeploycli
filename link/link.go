package link

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"io"

	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

// Machine structure
type Machine struct {
	Project     projects.Project
	Links       []*Link
	ErrStream   io.Writer
	Errors      *Errors
	ErrorsMutex sync.Mutex
	dirMutex    sync.Mutex
	queue       sync.WaitGroup
	list        *list.List
	RWList      list.RestartWatchList
	renameSID   []RenameServiceID
	end         bool
	endMutex    sync.RWMutex
}

// RenameServiceID object
type RenameServiceID struct {
	Any  bool
	From string
	To   string
}

// Match rule for renaming
func (r RenameServiceID) Match(serviceID string) bool {
	return r.Any || serviceID == r.From
}

// Link holds the information of service to be linked
type Link struct {
	Service     *services.Service
	ServicePath string
}

// ServiceError struct
type ServiceError struct {
	ServicePath string
	Error         error
}

// Errors list
type Errors struct {
	List []ServiceError
}

// New Service link
func New(dir string) (*Link, error) {
	return new(dir)
}

func new(dir string, renames ...RenameServiceID) (*Link, error) {
	var l = &Link{
		ServicePath: dir,
	}

	cp, err := services.Read(l.ServicePath)

	if err != nil {
		return nil, err
	}

	l.Service = cp.Service()
	maybeRenameService(l.Service, renames)

	verbose.Debug(fmt.Sprintf("Service ServiceID %v (local: %v) for directory %v",
		l.Service.ServiceID,
		cp.ID,
		dir))

	return l, err
}

func maybeRenameService(c *services.Service, renames []RenameServiceID) {
	for _, r := range renames {
		if r.Match(c.ServiceID) {
			c.ServiceID = r.To
			return
		}
	}
}

func (le Errors) Error() string {
	var msgs = []string{}

	for _, e := range le.List {
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ServicePath, e.Error.Error()))
	}

	return fmt.Sprintf("Local deployment errors:\n%v", strings.Join(msgs, "\n"))
}

var errMissingProjectID = errors.New("Missing project ID for linking services")

// Setup the linking machine
func (m *Machine) Setup(servicesList []string, renameServiceID ...RenameServiceID) error {
	if m.Project.ProjectID == "" {
		return errMissingProjectID
	}

	if err := m.initializeHealthUIDTable(); err != nil {
		return err
	}

	m.Errors = &Errors{
		List: []ServiceError{},
	}

	m.renameSID = renameServiceID

	for _, dir := range servicesList {
		m.mount(dir)
	}

	return nil
}

func (m *Machine) initializeHealthUIDTable() error {
	m.RWList.SetInitialProjectHealthUID(m.Project.HealthUID)

	var existingServices, err = m.Project.Services(context.Background())

	if err != nil {
		return err
	}

	var mt = map[string]string{}

	for _, s := range existingServices {
		mt[s.ServiceID] = s.HealthUID
	}

	m.RWList.SetInitialServicesHealthUID(mt)
	return nil
}

// Watch changes due to linking
func (m *Machine) Watch() {
	var cs []string

	for _, l := range m.Links {
		cs = append(cs, l.Service.ServiceID)
	}

	m.RWList = list.RestartWatchList{
		Project:        m.Project.ProjectID,
		Services:     cs,
		IsStillRunning: m.hasLinkingFinished,
	}

	m.RWList.Watch()
}

func (m *Machine) hasLinkingFinished() bool {
	m.endMutex.RLock()
	defer m.endMutex.RUnlock()
	return m.end
}

// Run links the services of the list input
func (m *Machine) Run(cancel func()) {
	m.queue.Add(len(m.Links))
	m.linkAll()
	m.queue.Wait()
	m.endMutex.Lock()
	m.end = true
	m.endMutex.Unlock()
	cancel()
}

func (m *Machine) linkAll() {
	for _, cl := range m.Links {
		// this might be in parallel (with a goroutine call)
		// but currently there is an issue on the backend
		// that makes things fail randomly
		// see https://github.com/wedeploy/cli/issues/170
		m.doLink(cl)
	}
}

func (m *Machine) doLink(cl *Link) {
	var err = m.link(cl)

	switch err {
	case nil:
		if m.ErrStream != nil {
			fmt.Fprintf(m.ErrStream, "Service %v deployed locally.\n", cl.Service.ServiceID)
		}
	default:
		m.logError(cl.ServicePath, err)
	}

	m.queue.Done()
}

func (m *Machine) link(l *Link) error {
	m.dirMutex.Lock()
	var err = services.Link(context.Background(),
		m.Project.ProjectID,
		*l.Service,
		l.ServicePath)
	m.dirMutex.Unlock()
	runtime.Gosched()

	return err
}

func (m *Machine) logError(dir string, err error) {
	m.ErrorsMutex.Lock()
	m.Errors.List = append(m.Errors.List, ServiceError{
		ServicePath: dir,
		Error:         errorhandling.Handle(err),
	})

	m.ErrorsMutex.Unlock()
}

func (m *Machine) mount(dir string) {
	var l, err = new(dir, m.renameSID...)

	if err != nil {
		m.logError(dir, err)
		return
	}

	m.Links = append(m.Links, l)
}
