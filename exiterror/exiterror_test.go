package exiterror

import (
	"testing"
)

func TestNew(t *testing.T) {
	var err error = New("this is an error", 67)
	want := "this is an error"
	code := 67

	if err == nil {
		t.Error("Expected error not to nil")
	}

	if err.Error() != want {
		t.Errorf("Wanted error message to be %v, got %v instead", want, err.Error())
	}

	e, ok := err.(Error)

	if !ok {
		t.Error("Expected error to be coerced to Error")
	}

	if e.Code() != code {
		t.Errorf("Expected exit code to be %v, got %v instead", e.Code(), code)
	}
}
