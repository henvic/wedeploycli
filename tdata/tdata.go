package tdata

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
)

// FromFile gets a string data from a file
func FromFile(filename string) string {
	b, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var value = string(b)

	if runtime.GOOS == "windows" {
		value = strings.Replace(value, "\r\n", "\n", -1)
	}

	return value
}

// ToFile is a helper function to create a file from a given content
func ToFile(filename string, content string) {
	var err = ioutil.WriteFile(filename, []byte(content), 0644)

	if err != nil {
		panic(err)
	}
}

// ServerHandler serves string content
func ServerHandler(content string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, content)
	}
}

// ServerJSONHandler serves string content as JSON
func ServerJSONHandler(content string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, content)
	}
}

// ServerFileHandler serves static content from a file
func ServerFileHandler(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, FromFile(filename))
	}
}

// ServerJSONFileHandler serves static JSON content from a file
func ServerJSONFileHandler(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, FromFile(filename))
	}
}
