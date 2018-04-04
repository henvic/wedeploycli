package socketio

import (
	"reflect"
	"testing"
)

func TestHandlerNotFunction(t *testing.T) {
	var notFunction []string

	var h, err = NewHandler(notFunction)

	if h != nil {
		t.Errorf("Expected no listener to be returned")
	}

	if err == nil || err.Error() != "listener is a slice instead of a function" {
		t.Errorf("Expected error doesn't match expected value, got %v", err)
	}

	switch err.(type) {
	case ErrorNotFunction:
	default:
		t.Errorf("Expected error to be of type NotFunctionError")
	}
}

func TestHandlerValidFunction(t *testing.T) {
	var _, err = NewHandler(mockHandlerStringInAndNumberOut)

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}
}

func TestParamKindInvalid(t *testing.T) {
	if err := checkParamKind(reflect.Func); err == nil {
		t.Errorf("Expected error when passing invalid parameter kind, got %v instead", err)
	}
}

func mockHandlerStringInAndNumberOut(s string) int {
	return len(s)
}
