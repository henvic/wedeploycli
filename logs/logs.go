package logs

import (
	"io"
	"os"
	"strconv"
	"time"
)

// LongPoolingPeriod is the time between retries
const LongPoolingPeriod = time.Second

// Logs structure
type Logs struct {
	AppName    string `json:"appName"`
	InstanceID string `json:"instanceId"`
	Level      int    `json:"level"`
	Message    string `json:"message"`
	PodName    string `json:"podName"`
	Severity   string `json:"severity"`
	Timestamp  string `json:"timestamp"`
}

// Filter structure
type Filter struct {
	Level      int    `json:"level,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	Since      string `json:"start,omitempty"`
}

// SeverityToLevel map
var SeverityToLevel = map[string]int{
	"critical": 2,
	"error":    3,
	"warning":  4,
	"info":     6,
	"debug":    7,
}

var outStream io.Writer = os.Stdout

// GetLevel to get level from severity or itself
func GetLevel(severityOrLevel string) (int, error) {
	var level = SeverityToLevel[severityOrLevel]

	if level != 0 {
		return level, nil
	}

	if severityOrLevel == "" {
		return 0, nil
	}

	return strconv.Atoi(severityOrLevel)
}
