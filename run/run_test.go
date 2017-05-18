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
	var tcpPorts = tcpPortsMap{
		TCPPort{
			Internal: 80,
			Expose:   80,
		},
		TCPPort{
			Internal: 8000,
			Expose:   8000,
		},
		TCPPort{
			Internal: 9001,
			Expose:   9002,
		},
	}

	var de = []string{
		"--publish", "80:80",
		"--publish", "8000:8000",
		"--publish", "9002:9001",
	}

	if !reflect.DeepEqual(tcpPorts.expose(), de) {
		t.Errorf("Expected ports exposure doesn't match expected value")
	}
}

func TestTCPPortsAvailableNone(t *testing.T) {
	var tcpPorts = tcpPortsMap{}

	var all, notAvailable = tcpPorts.getAvailability()

	if !all {
		t.Errorf("Availability should be true if no ports are required")
	}

	if len(notAvailable) != 0 {
		t.Errorf("Expected notAvailable to have length 0, got %v instead", notAvailable)
	}
}

func TestTCPPortsAvailableNotFree(t *testing.T) {
	var l, e = net.ListenTCP("tcp", &net.TCPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})

	if e != nil {
		panic(e)
	}

	var add = strings.SplitAfter(l.Addr().String(), ":")[1]
	var port, ea = strconv.Atoi(add)

	if ea != nil {
		panic(ea)
	}

	var tcpPorts = tcpPortsMap{
		TCPPort{
			Internal: port,
			Expose:   port,
		},
	}

	var all, notAvailable = tcpPorts.getAvailability()

	if all {
		t.Errorf("Availability should be false because port %v is in use", port)
	}

	var expectedNotavailable = []int{port}

	if !reflect.DeepEqual(notAvailable, expectedNotavailable) {
		t.Errorf("Not available arrays should be equal")
	}

	if err := l.Close(); err != nil {
		panic(err)
	}
}

func TestTCPPortsAvailableFree(t *testing.T) {
	var l, e = net.ListenTCP("tcp", &net.TCPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})

	if e != nil {
		panic(e)
	}

	var add = strings.SplitAfter(l.Addr().String(), ":")[1]
	var port, ea = strconv.Atoi(add)

	if ea != nil {
		panic(ea)
	}

	if err := l.Close(); err != nil {
		panic(err)
	}

	var tcpPorts = tcpPortsMap{
		TCPPort{
			Internal: port,
			Expose:   port,
		},
	}

	var all, notAvailable = tcpPorts.getAvailability()

	if !all {
		t.Errorf("Availability should be true because port %v is freed", port)
	}

	if len(notAvailable) != 0 {
		t.Errorf("There should be no non-available TCP ports, got %v instead", notAvailable)
	}
}
