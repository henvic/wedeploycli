package remoteuriparser

import (
	"net"
	"strings"
)

const (
	remoteAPIPrefix = "api.dashboard."
	defaultScheme   = "http"
)

// Parse address for extracting remote API
func Parse(address string) string {
	address = strings.TrimSuffix(address, "/")

	if isIP(address) {
		return defaultScheme + "://" + address
	}

	split := strings.SplitAfterN(address, "://", 2)

	if len(split) == 2 {
		return split[0] + addAPIEndpoint(split[1])
	}

	return defaultScheme + "://" + addAPIEndpoint(split[0])
}

func addAPIEndpoint(s string) string {
	if isIP(s) {
		return s
	}

	return remoteAPIPrefix + s
}

func isIP(host string) bool {
	h, _, err := net.SplitHostPort(host)

	if err == nil {
		host = h
	}

	return net.ParseIP(host) != nil
}
