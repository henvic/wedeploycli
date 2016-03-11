package prompt

import (
	"bytes"
	"os"
	"testing"
)

var bufInStream bytes.Buffer
var bufErrStream bytes.Buffer
var bufOutStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultInStream = inStream
	var defaultErrStream = errStream
	var defaultOutStream = outStream
	inStream = &bufInStream
	errStream = &bufErrStream
	outStream = &bufOutStream
	ec := m.Run()
	inStream = defaultInStream
	errStream = defaultErrStream
	outStream = defaultOutStream
	os.Exit(ec)
}

func TestIsSecretKey(t *testing.T) {
	originalSecretKeys := secretKeys
	secretKeys = []string{
		"token",
	}

	if !(isSecretKey("token") && isSecretKey("Token") && isSecretKey("TOKEN")) {
		t.Error(`isSecretKey failed to assert as secret key "token"`)
	}

	if isSecretKey("x") {
		t.Error(`isSecretKey failed to assert as non-secret key "x"`)
	}

	secretKeys = originalSecretKeys
}

func TestPrompt(t *testing.T) {
	var want = "value"
	bufInStream.Reset()
	bufInStream.WriteString("value")

	var u = Prompt("question")

	if u != want {
		t.Errorf("Expected prompt value %v, got %v instead", want, u)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "question: " {
		t.Error("Unexpected output stream")
	}
}
