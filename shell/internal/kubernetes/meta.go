// Package kubernetes has objects from k8s.io/apimachinery/pkg/apis/meta/v1/types.go
// https://github.com/kubernetes/apimachinery/blob/7ae370969693753028c74c0c876eee17ddb1b15b/pkg/apis/meta/v1/types.go
package kubernetes

// StatusDetails is a set of properties that might contain data regarding a command execution.
type StatusDetails struct {
	Causes []StatusCause `json:"causes,omitempty"`
}

// StatusCause provides more information about an api.Status failure, including
// cases when multiple errors are encountered.
type StatusCause struct {
	Type    string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}
