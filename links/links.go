package links

import "sync"

// Link to expose.
type Link struct {
	WeDeploy string
	DXP      string
}

// String gets the link related to the current infrastructure.
func (l Link) String() string {
	defer m.RUnlock()
	m.RLock()

	if dxp {
		return l.DXP
	}

	return l.WeDeploy
}

var dxp = false
var m sync.RWMutex

// SetDXP sets the link option to DXP.
func SetDXP() {
	m.Lock()
	defer m.Unlock()
	dxp = true
}

// SetWeDeploy sets the link option to WeDeploy.
func SetWeDeploy() {
	m.Lock()
	defer m.Unlock()
	dxp = false
}
