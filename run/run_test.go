package run

import (
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type ExistsDependencyProvider struct {
	cmd  string
	find bool
}

var ExistsDependencyCases = []ExistsDependencyProvider{
	{"git", true},
	{fmt.Sprintf("not-found-%d", rand.Int()), false},
}

func TestExistsDependency(t *testing.T) {
	for _, c := range ExistsDependencyCases {
		exists := existsDependency(c.cmd)

		if exists != c.find {
			t.Errorf("existsDependency(%v) should return %v", c.cmd, c.find)
		}
	}
}

func TestTCPPortsExpose(t *testing.T) {
	var originalTCPPorts = tcpPorts
	tcpPorts = tcpPortsStruct{80, 8000, 9000}
	var de = []string{
		"-p", "80:80",
		"-p", "8000:8000",
		"-p", "9000:9000",
	}

	if !reflect.DeepEqual(tcpPorts.expose(), de) {
		t.Errorf("Expected ports exposure doesn't match expected value")
	}

	tcpPorts = originalTCPPorts
}

func TestTCPPortsAvailableNone(t *testing.T) {
	var originalTCPPorts = tcpPorts
	tcpPorts = tcpPortsStruct{}

	if !tcpPorts.areAvailable() {
		t.Errorf("Availability should be true if no ports are required")
	}

	tcpPorts = originalTCPPorts
}

func TestTCPPortsAvailableNotFree(t *testing.T) {
	var l, e = net.ListenTCP("tcp", &net.TCPAddr{})

	if e != nil {
		panic(e)
	}

	var add = strings.SplitAfter(l.Addr().String(), "[::]:")[1]
	var port, ea = strconv.Atoi(add)

	if ea != nil {
		panic(ea)
	}

	var originalTCPPorts = tcpPorts
	tcpPorts = tcpPortsStruct{port}

	if tcpPorts.areAvailable() {
		t.Errorf("Availability should be false because port %v is in use", port)
	}

	l.Close()
	tcpPorts = originalTCPPorts
}

func TestTCPPortsAvailableFree(t *testing.T) {
	var l, e = net.ListenTCP("tcp", &net.TCPAddr{})

	if e != nil {
		panic(e)
	}

	var add = strings.SplitAfter(l.Addr().String(), "[::]:")[1]
	var port, ea = strconv.Atoi(add)

	if ea != nil {
		panic(ea)
	}

	l.Close()

	var originalTCPPorts = tcpPorts
	tcpPorts = tcpPortsStruct{port}

	if !tcpPorts.areAvailable() {
		t.Errorf("Availability should be true because port %v is freed", port)
	}

	tcpPorts = originalTCPPorts
}
