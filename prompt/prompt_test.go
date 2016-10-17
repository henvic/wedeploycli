package prompt

import (
	"bytes"
	"os"
	"testing"
)

var (
	bufInStream       bytes.Buffer
	bufErrStream      bytes.Buffer
	bufOutStream      bytes.Buffer
	defaultIsTerminal = isTerminal
)

func TestMain(m *testing.M) {
	var defaultInStream = inStream
	var defaultErrStream = errStream
	var defaultOutStream = outStream
	inStream = &bufInStream
	errStream = &bufErrStream
	outStream = &bufOutStream
	ec := m.Run()
	isTerminal = defaultIsTerminal
	inStream = defaultInStream
	errStream = defaultErrStream
	outStream = defaultOutStream
	os.Exit(ec)
}

func TestSelectOption(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("2\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, nil)

	if option != 1 {
		t.Errorf("Expected option to be 1 (index for 2), got %v instead", option)
	}

	if errt != nil {
		t.Errorf("Expected option error to be nil, got %v instead", errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "\nSelect from 1..4: " {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionEquivalentChoosen(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("pass2\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, map[string]int{
		"fail1": 1,
		"pass2": 2,
		"fail3": 3,
		"fail4": 4,
	})

	if option != 1 {
		t.Errorf("Expected option to be 1 (index for 2), got %v instead", option)
	}

	if errt != nil {
		t.Errorf("Expected option error to be nil, got %v instead", errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "\nSelect from 1..4: " {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionEquivalentNotChoosen(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("2\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, map[string]int{
		"fail1": 1,
		"pass2": 2,
		"fail3": 3,
		"fail4": 4,
	})

	if option != 1 {
		t.Errorf("Expected option to be 1 (index for 2), got %v instead", option)
	}

	if errt != nil {
		t.Errorf("Expected option error to be nil, got %v instead", errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "\nSelect from 1..4: " {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionIsNotTerminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(5, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "Input device is not a terminal. Can't read \"\nSelect from 1..5\""

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "" {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionNoneAvailable(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(0, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "No options available."

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "" {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionNoneAvailableEquivalent(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(0, map[string]int{
		"foo": 1,
	})

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "No options available."

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "" {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionInvalidOption(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "Invalid option."

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "\nSelect from 1..4: " {
		t.Error("Unexpected output stream")
	}
}

func TestSelectOptionInvalidOptionEquivalent(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, map[string]int{
		"foo": 1,
		"bar": 2,
	})

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "Invalid option."

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "\nSelect from 1..4: " {
		t.Error("Unexpected output stream")
	}
}

func TestPrompt(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	var want = "value"
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Prompt("question")

	if errt != nil {
		t.Errorf("Expected prompt error to be nil, got %v instead", errt)
	}

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

func TestPromptWithSpace(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	var want = "my value"
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = true
	_, _ = bufInStream.WriteString("my value\n")

	var u, errt = Prompt("question")

	if errt != nil {
		t.Errorf("Expected prompt error to be nil, got %v instead", errt)
	}

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

func TestPromptIsNotterminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Prompt("question")

	if errt == nil {
		t.Errorf("Expected prompt error to be not nil, got %v instead", errt)
	}

	var wantErr = `Input device is not a terminal. Can't read "question"`

	if errt.Error() != wantErr {
		t.Errorf("Expected error message %v, got %v instead", wantErr, errt)
	}

	if u != "" {
		t.Errorf("Expected prompt value empty, got %v instead", u)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "" {
		t.Error("Unexpected output stream")
	}
}

func TestHiddenIsNotterminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	bufErrStream.Reset()
	bufOutStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Hidden("question")

	if errt == nil {
		t.Errorf("Expected prompt error to be not nil, got %v instead", errt)
	}

	var wantErr = `Input device is not a terminal. Can't read "question"`

	if errt.Error() != wantErr {
		t.Errorf("Expected error message %v, got %v instead", wantErr, errt)
	}

	if u != "" {
		t.Errorf("Expected prompt value empty, got %v instead", u)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Expected error stream to be empty")
	}

	if bufOutStream.String() != "" {
		t.Error("Unexpected output stream")
	}
}
