package configstore

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/launchpad-project/api.go/jsonlib"
)

func TestLoad(t *testing.T) {
	var s = Store{
		Name: "myconfig",
		Path: "./mocks/example.json",
	}

	if err := s.Load(); err != nil {
		t.Errorf("Unexpected error %s, when no error should happen", err)
	}

	jsonlib.AssertJSONMarshal(t, `{"name": "helloworld"}`, s.Data)
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

	if w, err := s.GetString("planet.Earth.BR.city"); w != "Recife" || err != nil {
		t.Errorf("Wanted <Recife, nil>, got %s %s instead", w, err)
	}

	if w, err := s.GetString("score"); w != "1000" || err != nil {
		t.Errorf("Wanted score = (string) 1000, got <%s, %s> instead", w, err)
	}

	if w, err := s.GetString("sys.components"); w != "null" || err != nil {
		t.Errorf("Wanted sys.components = null, got <%s, %s> instead", w, err)
	}

	if w, err := s.GetString("sys.missing"); w != "" || err != ErrConfigKeyNotFound {
		t.Errorf("Wanted sys.missing to be <\"\", ErrConfigKeyNotFound>, got <%s, %s> instead", w, err)
	}

	if w, err := s.GetString(""); w != "" || err != ErrConfigKeyNotFound {
		t.Errorf("Wanted (empty) to be <\"\", ErrConfigKeyNotFound>, got <%s, %s> instead", w, err)
	}

	if w, err := s.GetString("..."); w != "" || err != ErrConfigKeyNotFound {
		t.Errorf("Wanted ... to be <\"\", ErrConfigKeyNotFound>, got <%s, %s> instead", w, err)
	}

	if w, err := s.GetString("planet"); w != "" || err != ErrConfigKeyNotLeaf {
		t.Errorf("Wanted planet to be <empty, %s>, got <%s, %s> instead", ErrConfigKeyNotLeaf, w, err)
	}
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

func TestSave(t *testing.T) {
	os.Remove("./mocks/temporary.json")

	var s = Store{
		Name: "myconfig",
		Path: "./mocks/temporary.json",
	}

	s.SetAndSave("foo.bah", "hello")
}

func TestSet(t *testing.T) {
	var s = Store{}

	s.Set("foo.bah", "value")
	jsonlib.AssertJSONMarshal(t, `{"foo":{"bah": "value"}}`, s.Data)
}

func TestSetEditableKey(t *testing.T) {
	var s = Store{
		ConfigurableKeys: map[string]bool{
			"user.account.token": true,
		},
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
		Name: "myconfig",
		Path: "./mocks/temporary.json",
		ConfigurableKeys: map[string]bool{
			"foo.bah": true,
		},
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
