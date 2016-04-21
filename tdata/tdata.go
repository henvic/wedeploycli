package tdata

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// FromFile gets a string data from a file
func FromFile(filename string) string {
	b, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	return string(b)
}

// ServerHandler serves string content
func ServerHandler(content string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
