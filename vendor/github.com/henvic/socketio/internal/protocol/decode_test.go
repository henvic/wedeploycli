package protocol

import (
	"reflect"
	"testing"
)

func TestDecodeEmptyMessage(t *testing.T) {
	m, err := Decode([]byte(""))

	if m != nil {
		t.Errorf("Expected message to be nil, got %v instead", m)
	}

	if err != ErrorWrongMessageType {
		t.Errorf("Expected error to be %v, got %v instead", ErrorWrongMessageType, err)
	}
}

func TestDecodeMessageWithNamespace(t *testing.T) {
	m, err := Decode([]byte(`42/shell,["stdout", "$ ls"]`))

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	want := Message{
		Namespace: "/shell",
		Method:    "stdout",
		Type:      MessageTypeEmit,
		Data:      []byte(`[ "$ ls"]`),
		Source:    `42/shell,["stdout", "$ ls"]`,
	}

	if !reflect.DeepEqual(want, *m) {
		t.Errorf("Expected %+v to match %+v", m, want)
	}
}

func TestDecodeEmptyMessageWithNamespace(t *testing.T) {
	m, err := Decode([]byte(`40/subscribe/project/service/container,`))

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	if m.Type != MessageTypeEmpty {
		t.Errorf("Expected type to be %v, got %v instead", MessageTypeEmpty, m.Type)
	}

	wantNamespace := "/subscribe/project/service/container"

	if m.Namespace != wantNamespace {
		t.Errorf("Expected type to be %v, got %v instead", wantNamespace, m.Namespace)
	}
}
