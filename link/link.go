package link

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"io"

	"github.com/wedeploy/cli/containers"
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

// Link holds the information of container to be linked
type Link struct {
	Container     *containers.Container
	ContainerPath string
}

// ContainerError struct
type ContainerError struct {
	ContainerPath string
	Error         error
}

// Errors list
type Errors struct {
	List []ContainerError
}

// New Container link
func New(dir string) (*Link, error) {
	return new(dir)
}

func new(dir string, renames ...RenameServiceID) (*Link, error) {
	var l = &Link{
		ContainerPath: dir,
	}

	cp, err := containers.Read(l.ContainerPath)

	if err != nil {
		return nil, err
	}

	l.Container = cp.Container()
	maybeRenameContainer(l.Container, renames)

	verbose.Debug(fmt.Sprintf("Container ServiceID %v (local: %v) for directory %v",
		l.Container.ServiceID,
		cp.ID,
		dir))

	return l, err
}

func maybeRenameContainer(c *containers.Container, renames []RenameServiceID) {
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
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ContainerPath, e.Error.Error()))
	}

	return fmt.Sprintf("Local deployment errors:\n%v", strings.Join(msgs, "\n"))
}

var errMissingProjectID = errors.New("Missing project ID for linking containers")

// Setup the linking machine
func (m *Machine) Setup(containersList []string, renameServiceID ...RenameServiceID) error {
	if m.Project.ProjectID == "" {
		return errMissingProjectID
	}

	if err := m.initializeHealthUIDTable(); err != nil {
		return err
	}

	m.Errors = &Errors{
		List: []ContainerError{},
	}

	m.renameSID = renameServiceID

	for _, dir := range containersList {
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

	m.RWList.SetInitialContainersHealthUID(mt)
	return nil
}

// Watch changes due to linking
func (m *Machine) Watch() {
	var cs []string

	for _, l := range m.Links {
		cs = append(cs, l.Container.ServiceID)
	}

	m.RWList = list.RestartWatchList{
		Project:        m.Project.ProjectID,
		Containers:     cs,
		IsStillRunning: m.hasLinkingFinished,
	}

	m.RWList.Watch()
}

func (m *Machine) hasLinkingFinished() bool {
	m.endMutex.RLock()
	defer m.endMutex.RUnlock()
	return m.end
}

// Run links the containers of the list input
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
			fmt.Fprintf(m.ErrStream, "Container %v deployed locally.\n", cl.Container.ServiceID)
		}
	default:
		m.logError(cl.ContainerPath, err)
	}

	m.queue.Done()
}

func (m *Machine) link(l *Link) error {
	m.dirMutex.Lock()
	var err = containers.Link(context.Background(),
		m.Project.ProjectID,
		*l.Container,
		l.ContainerPath)
	m.dirMutex.Unlock()
	runtime.Gosched()

	return err
}

func (m *Machine) logError(dir string, err error) {
	m.ErrorsMutex.Lock()
	m.Errors.List = append(m.Errors.List, ContainerError{
		ContainerPath: dir,
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
