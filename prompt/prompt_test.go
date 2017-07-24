package prompt

import (
	"bytes"
	"os"
	"testing"
)

var (
	bufInStream       bytes.Buffer
	defaultIsTerminal = isTerminal
)

func TestMain(m *testing.M) {
	var defaultInStream = inStream
	inStream = &bufInStream
	ec := m.Run()
	isTerminal = defaultIsTerminal
	inStream = defaultInStream
	os.Exit(ec)
}

func TestSelectOption(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
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
}

func TestSelectOptionEquivalentChoosen(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
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
}

func TestSelectOptionEquivalentNotChoosen(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
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
}

func TestSelectOptionIsNotTerminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(5, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "input device is not a terminal"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestSelectOptionNoneAvailable(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(0, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "no options available"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestSelectOptionNoneAvailableEquivalent(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
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

	var wantErr = "no options available"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestSelectOptionInvalidOption(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "invalid option"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestSelectOptionInvalidOptionOffByOne(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("5\n")

	if err != nil {
		panic(err)
	}

	var option, errt = SelectOption(4, nil)

	if option != -1 {
		t.Errorf("Expected option to be -1, got %v instead", option)
	}

	var wantErr = "invalid option"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestSelectOptionInvalidOptionEquivalent(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
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

	var wantErr = "invalid option"

	if errt == nil || errt.Error() != wantErr {
		t.Errorf("Expected option error to be %v, got %v instead", wantErr, errt)
	}
}

func TestPrompt(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	var want = "value"
	bufInStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Prompt()

	if errt != nil {
		t.Errorf("Expected prompt error to be nil, got %v instead", errt)
	}

	if u != want {
		t.Errorf("Expected prompt value %v, got %v instead", want, u)
	}
}

func TestPromptEmpty(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	var want = ""
	bufInStream.Reset()
	isTerminal = true
	var _, err = bufInStream.WriteString("\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Prompt()

	if errt != nil {
		t.Errorf("Expected prompt error to be nil, got %v instead", errt)
	}

	if u != want {
		t.Errorf("Expected prompt value %v, got %v instead", want, u)
	}
}

func TestPromptWithSpace(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	var want = "my value"
	bufInStream.Reset()
	isTerminal = true
	_, _ = bufInStream.WriteString("my value\n")

	var u, errt = Prompt()

	if errt != nil {
		t.Errorf("Expected prompt error to be nil, got %v instead", errt)
	}

	if u != want {
		t.Errorf("Expected prompt value %v, got %v instead", want, u)
	}
}

func TestPromptIsNotterminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Prompt()

	if errt == nil {
		t.Errorf("Expected prompt error to be not nil, got %v instead", errt)
	}

	var wantErr = `input device is not a terminal`

	if errt.Error() != wantErr {
		t.Errorf("Expected error message %v, got %v instead", wantErr, errt)
	}

	if u != "" {
		t.Errorf("Expected prompt value empty, got %v instead", u)
	}
}

func TestHiddenIsNotterminal(t *testing.T) {
	defer func() {
		isTerminal = defaultIsTerminal
	}()
	bufInStream.Reset()
	isTerminal = false
	var _, err = bufInStream.WriteString("value\n")

	if err != nil {
		panic(err)
	}

	var u, errt = Hidden()

	if errt == nil {
		t.Errorf("Expected prompt error to be not nil, got %v instead", errt)
	}

	var wantErr = `input device is not a terminal: can't read password`

	if errt.Error() != wantErr {
		t.Errorf("Expected error message %v, got %v instead", wantErr, errt)
	}

	if u != "" {
		t.Errorf("Expected prompt value empty, got %v instead", u)
	}
}
