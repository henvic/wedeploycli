package configstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Store is the structure for the config object
type Store struct {
	Name string
	Path string
	// ConfigurableKeys must be configurable string values only
	ConfigurableKeys []string
	Data             map[string]interface{}
}

var (
	// ErrConfigKeyNotFound for when a key is not found on the config
	ErrConfigKeyNotFound = errors.New("key not found")

	// ErrConfigKeyNotLeaf for when a non-leaf key is found on the config
	ErrConfigKeyNotLeaf = errors.New("key not leaf")

	// ErrConfigKeyNotConfigurable is used when a key is not configurable
	ErrConfigKeyNotConfigurable = errors.New("key not configurable")

	outStream io.Writer = os.Stdout
)

// Load reads and populates the config data
func (s *Store) Load() error {
	content, err := ioutil.ReadFile(s.Path)

	if err == nil {
		err = json.Unmarshal(content, &s.Data)
	}

	return err
}

// Save the config to the config file on Path
func (s *Store) Save() {
	bin, err := json.MarshalIndent(s.Data, "", "    ")

	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(s.Path, bin, 0644); err != nil {
		panic(err)
	}
}

// Get a string or return empty when not found
func (s *Store) Get(key string) string {
	value, _ := s.GetString(key)
	return value
}

// GetInterface gets an interface from the config object
func (s *Store) GetInterface(key string) (interface{}, error) {
	var keyPath = strings.Split(key, ".")
	var parent = s.Data

	for pos, subPath := range keyPath {
		if _, exists := parent[subPath]; !exists {
			break
		}

		if pos == len(keyPath)-1 {
			return parent[subPath], nil
		}

		parent = parent[subPath].(map[string]interface{})
	}

	return "", ErrConfigKeyNotFound
}

// GetRequiredString and panics when not found
func (s *Store) GetRequiredString(key string) string {
	value, err := s.GetString(key)

	if err != nil {
		panic(err)
	}

	return value
}

// GetString from the config
func (s *Store) GetString(key string) (string, error) {
	var keyPath = strings.Split(key, ".")
	var parent = s.Data
	var subPath string
	var pos int

	for pos, subPath = range keyPath {
		if _, exists := parent[subPath]; !exists {
			return "", ErrConfigKeyNotFound
		}

		if pos == len(keyPath)-1 {
			break
		}

		parent = parent[subPath].(map[string]interface{})
	}

	switch parent[subPath].(type) {
	case nil:
		return "null", nil
	case string, int, int64, float64, bool:
		return fmt.Sprintf("%v", parent[subPath]), nil
	default:
		return "", ErrConfigKeyNotLeaf
	}
}

// List configurable keys
func (s *Store) List() {
	for _, key := range s.ConfigurableKeys {
		fmt.Fprintln(outStream, key+" = "+s.GetRequiredString(key))
	}
}

// Set sets the value for a given key
func (s *Store) Set(key, value string) {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}

	var keyPath = strings.Split(key, ".")
	var parent = s.Data

	for pos, subPath := range keyPath {
		if pos == len(keyPath)-1 {
			parent[subPath] = value
			continue
		}

		switch parent[subPath].(type) {
		case map[string]interface{}:
		default:
			parent[subPath] = make(map[string]interface{})
		}

		parent = parent[subPath].(map[string]interface{})
	}
}

// SetAndSave sets the value for a given key and save config
func (s *Store) SetAndSave(key, value string) {
	s.Set(key, value)
	s.Save()
}

// SetAndSaveEditableKey sets the value for a given key and save config
// (or fail if it the key is not public)
func (s *Store) SetAndSaveEditableKey(key, value string) error {
	if err := s.SetEditableKey(key, value); err != nil {
		return err
	}

	s.Save()
	return nil
}

// SetEditableKey sets the value for a given key
// (or fail if it the key is not public)
func (s *Store) SetEditableKey(key, value string) error {
	for _, v := range s.ConfigurableKeys {
		if v == key {
			s.Set(key, value)
			return nil
		}
	}

	return ErrConfigKeyNotConfigurable
}
