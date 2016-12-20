package collector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"bytes"

	"strings"

	uuid "github.com/satori/go.uuid"
	wedeploy "github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/verbosereq"
)

var (
	// Debug mode
	Debug bool

	// Backend for the collector
	Backend string

	// Index for the collection
	Index = "we-cli"

	// Type for the collection
	Type = "metrics"

	// BulkSize tells how many objects the bulk collects
	BulkSize = 50

	// BackendTimeout for backend requests
	BackendTimeout = 5 * time.Second
)

func verboseF(format string, a ...interface{}) {
	if Debug {
		log.Printf(format, a...)
	}
}

// Handler for the collector requests
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed: metrics collector only accepts POST requests\n")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var host, _, _ = net.SplitHostPort(r.RemoteAddr)
	var bulkCollector = &BulkCollector{
		RequestID:     uuid.NewV4().String(),
		IP:            host,
		XForwardedFor: r.Header.Get("X-Forwarded-For"),
		Feedback: BulkCollectorFeedback{
			JSONFailureLines: []int{},
		},
		reader:     bufio.NewReader(r.Body),
		collection: [][]byte{},
		idToLine:   map[string]int{},
	}

	if err := bulkCollector.Run(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error trying to store analytics.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v", bulkCollector.GetFeedback())
}

// Event entry for collector
// see also metrics.Event
type Event struct {
	ID            string            `json:"id"`
	Type          string            `json:"event_type,omitempty"`
	Text          string            `json:"text,omitempty"`
	PID           string            `json:"pid,omitempty"`
	SID           string            `json:"sid,omitempty"`
	Time          string            `json:"time,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
	Scope         string            `json:"scope,omitempty"`
	Version       string            `json:"version,omitempty"`
	OS            string            `json:"os,omitempty"`
	Arch          string            `json:"arch,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	IP            string            `json:"ip,omitempty"`
	XForwardedFor string            `json:"x_forwarded_for,omitempty"`
	Raw           string            `json:"raw"`
}

// BulkCollector is used for collecting the lines and posting to the backend
type BulkCollector struct {
	RequestID     string
	IP            string
	XForwardedFor string
	Feedback      BulkCollectorFeedback
	reader        *bufio.Reader
	collection    [][]byte
	lineAc        int
	idToLine      map[string]int
	request       *wedeploy.WeDeploy
}

// BulkCollectorFeedback is used for printing the statistics of the request
type BulkCollectorFeedback struct {
	Errors           bool        `json:"errors"`
	Objects          int         `json:"objects"`
	JSONFailureLines []int       `json:"json_failure_lines"`
	Insertions       []Insertion `json:"insertions"`
}

// Run bulk collector
func (b *BulkCollector) Run() (err error) {
	for err != io.EOF {
		if err = b.tryToCollect(); err != nil && err != io.EOF {
			return err
		}
	}

	verboseF("Closing request %v\n", b.RequestID)
	return nil
}

type esResponseBulk struct {
	Errors bool                        `json:"errors"`
	Items  []map[string]esResponseItem `json:"items"`
}

type esResponseItem struct {
	ID     string `json:"_id"`
	Status int    `json:"status"`
}

// Insertion for insertion response
type Insertion struct {
	ID     string `json:"id"`
	Error  bool   `json:"error"`
	Line   int    `json:"line"`
	Status int    `json:"status"`
}

// GetFeedback for a bulk collection
func (b *BulkCollector) GetFeedback() string {
	var objResponses = esResponseBulk{}

	switch b.request {
	case nil:
		verboseF("Skipped backend request: no events to register were received\n")
	default:
		if err := apihelper.DecodeJSON(b.request, &objResponses); err != nil {
			log.Printf("Server response error: %v\n", err)
		}
	}

	b.Feedback.Errors = objResponses.Errors

	for _, o := range objResponses.Items {
		oc, ok := o["create"]

		var e = Insertion{
			ID:     oc.ID,
			Error:  oc.Status >= 400,
			Line:   b.idToLine[oc.ID],
			Status: oc.Status,
		}

		if !ok {
			e.Error = true
			e.Status = -1
		}

		b.Feedback.Insertions = append(b.Feedback.Insertions, e)
	}

	var j, err = json.Marshal(b.Feedback)

	if err != nil {
		log.Printf("Error marshaling feedback: %v", err)
	}

	return string(j)
}

func (b *BulkCollector) tryToCollect() error {
	var sb, err = b.reader.ReadBytes('\n')

	if err != nil && err != io.EOF {
		verboseF("Error reading from Request %v: %v\n", b.RequestID, err)
		return err
	}

	if len(sb) == 0 && len(b.collection) == 0 {
		return err
	}

	return b.collect(sb, err)
}

func (b *BulkCollector) collect(sb []byte, errRead error) (err error) {
	b.verboseCollect(sb)
	b.collection = append(b.collection, sb)

	if errRead == io.EOF || len(b.collection) >= BulkSize {
		if err = b.tryPostAndPop(); err != nil {
			return err
		}
	}

	return errRead
}

func (b *BulkCollector) verboseCollect(sb []byte) {
	rvf := fmt.Sprintf("Reading from %v", b.IP)

	if b.XForwardedFor != "" {
		rvf += fmt.Sprintf(" (X-Forwarded-For: %v)", b.XForwardedFor)
	}

	if Debug {
		verboseF("%v:\n%v", rvf, string(sb))
	}
}

func (b *BulkCollector) readObject(sb []byte, line int) (*Event, error) {
	var event = Event{}

	if len(sb) == 0 || len(sb) == 1 && string(sb) == "\n" {
		return nil, nil
	}

	if err := json.Unmarshal(sb, &event); err != nil {
		b.Feedback.JSONFailureLines = append(b.Feedback.JSONFailureLines, line+1)
		return &event, fmt.Errorf("Error trying to read object: %v", err)
	}

	return &event, nil
}

func getEventAction(id string) (s string) {
	// must be one-level for the ElasticSearch _bulk API to work
	var createTemplate = `{"create": {"_index": %v, "_type": %v, "_id": %v}}`

	var idField, _ = json.Marshal(id)
	var indexField, _ = json.Marshal(Index)
	var typeField, _ = json.Marshal(Type)
	return fmt.Sprintf(createTemplate, string(indexField), string(typeField), string(idField))
}

func (b *BulkCollector) process() *bytes.Buffer {
	var buffer = &bytes.Buffer{}

	for line, object := range b.collection {
		event, _ := b.readObject(object, b.lineAc+line)

		if event == nil {
			continue
		}

		if len(event.ID) == 0 {
			event.ID = uuid.NewV4().String()
		}

		b.idToLine[event.ID] = line + 1
		event.RequestID = b.RequestID
		event.IP = b.IP
		event.XForwardedFor = b.XForwardedFor
		event.Raw = strings.TrimSuffix(string(object), "\n")

		var j, err = json.Marshal(event)

		switch err {
		case nil:
			fmt.Fprintf(buffer, "%v\n%v\n", getEventAction(event.ID), string(j))
		default:
			log.Printf("Error trying to write bulk collector map: %v\n", err)
		}
	}

	return buffer
}

func (b *BulkCollector) tryPostAndPop() (err error) {
	var backoff = time.Duration(0)
	for try := 0; try < 3; try++ {
		if err = b.postAndPop(); err == nil {
			return nil
		}

		log.Printf("Error trying to POST to backend: %v\n", err)
		// wait a little bit before trying again
		backoff += (12 << uint(try)) * time.Millisecond
		time.Sleep(backoff)
	}

	return errors.New("Error communicating with the backend service")
}

func (b *BulkCollector) postAndPop() (err error) {
	buf := b.process()

	if buf.Len() == 0 {
		return nil
	}

	b.request = wedeploy.URL(Backend, "/events/_bulk")

	b.request.Body(buf)

	var ctx, cancelFunc = context.WithTimeout(context.Background(), BackendTimeout)
	b.request.SetContext(ctx)

	err = b.request.Post()
	cancelFunc()

	verbosereq.Feedback(b.request)

	if err != nil {
		return err
	}

	b.Feedback.Objects += len(b.collection)
	b.lineAc += len(b.collection)
	b.collection = [][]byte{}
	return nil
}
