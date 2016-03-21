package configstore

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/launchpad-project/api.go/jsonlib"
)

func assertGetString(t *testing.T, s Store, key, retVal string, retErr error) {
	if w, err := s.GetString(key); w != retVal || err != retErr {
		t.Errorf("Wanted <%s, %s>, got <%s %s> instead", retVal, retErr, w, err)
	}
}

func TestLoad(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/example.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %s, when no error should happen", err)
	}

	jsonlib.AssertJSONMarshal(t, `{
		"id": "helloworld",
		"name": "Hello World"
		}`, s.Data)
}

func TestLoadInvalidJSON(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/invalid.json",
	}

	var err = s.Load()

	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Expected syntax error for invalid JSON: got %T", err)
	}
}

func TestLoadNotFound(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/unknown.json",
	}

	var err = s.Load()

	if !os.IsNotExist(err) {
		t.Errorf("Unexpected error %s, when file not found was expected", err)
	}
}

func TestGet(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	if w := s.Get("planet.Earth.US.city"); w != "Diamond Bar" {
		t.Errorf("Wanted Diamond Bar, got %s instead", w)
	}

	if w := s.Get("unknown"); w != "" {
		t.Error("Wanted empty from unknown key, got something else instead")
	}
}

func TestGetReadingNonTerminal(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	defer func() {
		r := recover()

		if r != ErrConfigKeyNotLeaf {
			t.Errorf("Expected panic with %v error, got %v instead", ErrConfigKeyNotLeaf, r)
		}
	}()

	s.GetRequiredString("planet")
}

func TestGetRequiredString(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	if w := s.GetRequiredString("planet.Earth.US.city"); w != "Diamond Bar" {
		t.Errorf("Wanted Diamond Bar, got %s instead", w)
	}
}

func TestGetString(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	assertGetString(t, s, "planet.Earth.BR.city", "Recife", nil)
	assertGetString(t, s, "score", "1000", nil)
	assertGetString(t, s, "sys.components", "null", nil)
	assertGetString(t, s, "sys.missing", "", ErrConfigKeyNotFound)
	assertGetString(t, s, "", "", ErrConfigKeyNotFound)
	assertGetString(t, s, "...", "", ErrConfigKeyNotFound)
	assertGetString(t, s, "planet", "", ErrConfigKeyNotLeaf)
}

func TestGetInterface(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	var got, err = s.GetInterface("planet.Earth.BR")

	if err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	jsonlib.AssertJSONMarshal(t, `{"city": "Recife"}`, got)
}

func TestGetInterfaceFailure(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/complex.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	var got, err = s.GetInterface("planet.Earth.AR")

	if err != ErrConfigKeyNotFound {
		t.Errorf("Expected <%v, %v>, got <%v, %v> instead", nil, ErrConfigKeyNotFound, got, err)
	}
}

func TestList(t *testing.T) {
	var s = Store{
		Name:             "conf",
		Path:             "./mocks/read.json",
		ConfigurableKeys: []string{"id", "name", "sys.components", "score"},
	}

	var want = `id = complex
name = Complex
sys.components = null
score = 1000
`

	s.Load()

	var bufOutStream bytes.Buffer
	var defaultOutStream = outStream
	outStream = &bufOutStream

	s.List()

	var outString = bufOutStream.String()

	if outString != want {
		t.Errorf("Wanted config list %v, got %v instead", want, outString)
	}

	outStream = defaultOutStream
}

func TestSave(t *testing.T) {
	os.Remove("./mocks/temporary.json")

	var s = Store{
		Name: "myconfig",
		Path: "./mocks/temporary.json",
	}

	s.SetAndSave("foo.bah", "hello")
}

func TestSaveFailure(t *testing.T) {
	defer func() {
		r := recover()

		if !os.IsNotExist(r.(error)) {
			t.Errorf("Unexpected error %s, when file not found was expected", r)
		}
	}()

	var s = Store{
		Name: "myconfig",
		Path: "",
	}

	s.Save()
}

func TestSet(t *testing.T) {
	var s = Store{}

	s.Set("foo.bah", "value")
	jsonlib.AssertJSONMarshal(t, `{"foo":{"bah": "value"}}`, s.Data)
}

func TestSetEditableKey(t *testing.T) {
	var s = Store{
		ConfigurableKeys: []string{"user.account.token"},
	}
	var err = s.SetEditableKey("user.account.token", "value")

	if err != nil {
		t.Errorf("Expected error %v, got %v instead", nil, err)
	}
	jsonlib.AssertJSONMarshal(t, `{"user":{"account": {"token": "value"}}}`, s.Data)
}

func TestSetEditableKeyPanic(t *testing.T) {
	var s = Store{}
	var err = s.SetEditableKey("foo.bah", "value")

	if err != ErrConfigKeyNotConfigurable {
		t.Errorf("Expected error %v, got %v instead", ErrConfigKeyNotConfigurable, err)
	}

	_, err = s.GetString("foo.bah")

	if err != ErrConfigKeyNotFound {
		t.Errorf("Expected error %v, got %v instead", ErrConfigKeyNotFound, err)
	}
}

func TestSetFromJSON(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/read.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %v when no error was expected", err)
	}

	if w := s.GetRequiredString("planet.Earth.BR.city"); w != "Recife" {
		t.Errorf("Expected value to be Complex, got %v instead", w)
	}

	s.Set("planet.Earth.BR.city", "bah")

	if w := s.GetRequiredString("planet.Earth.BR.city"); w != "bah" {
		t.Errorf("Expected value to be bah, got %v instead", w)
	}
}

func TestSetAndSaveEditableKey(t *testing.T) {
	os.Remove("./mocks/temporary.json")

	var s = Store{
		Name:             "myconfig",
		Path:             "./mocks/temporary.json",
		ConfigurableKeys: []string{"foo.bah"},
	}

	if err := s.SetAndSaveEditableKey("foo.bah", "hello"); err != nil {
		t.Errorf("Expected foo.bah key to be configurable, got %v error instead", err)
	}
}

func TestSetAndSaveEditableKeyFailure(t *testing.T) {
	os.Remove("./mocks/temporary.json")

	var s = Store{
		Name: "myconfig",
		Path: "./mocks/temporary.json",
	}

	if err := s.SetAndSaveEditableKey("foo.bah", "hello"); err == nil {
		t.Errorf("Expected foo.bah key to be non-configurable, got %v error instead", err)
	}
}
