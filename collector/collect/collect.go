package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/wedeploy/cli/collector"
	"github.com/wedeploy/cli/verbose"
)

const (
	defaultAddr    = ":4884"
	defaultBackend = "http://localhost:9200/"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "addr", defaultAddr, "Serving address")
	flag.StringVar(&collector.Backend, "backend", defaultBackend, "ElasticSearch API address")
	flag.StringVar(&collector.Index, "index", collector.Index, "ElasticSearch Index")
	flag.StringVar(&collector.Type, "type", collector.Type, "ElasticSearch Type")
	flag.BoolVar(&collector.Debug, "debug", false, "Print debug messages and received objects")
	flag.Parse()
}

func main() {
	if collector.Backend == "" {
		fmt.Fprintf(os.Stderr, "-backend is not set\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Analytics collector started listening on %v\n", addr)

	if collector.Debug {
		fmt.Fprintf(os.Stderr, "Debug mode is enabled\n")
		verbose.Enabled = true
	}

	http.HandleFunc("/", collector.Handler)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
