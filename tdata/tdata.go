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

// ServerHandler serves static content from a file
func ServerHandler(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, FromFile(filename))
	}
}
