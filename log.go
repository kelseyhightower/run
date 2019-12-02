package run

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/kelseyhightower/run/metadata"
)

type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	component string
	out       io.Writer // destination for output
}

func NewLogger(component string) *Logger {
	return &Logger{out: os.Stdout, component: component}
}

func (l *Logger) Info(v ...interface{}) {
	l.Log("INFO", v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Log("ERROR", v...)
}

func (l *Logger) Notice(v ...interface{}) {
	l.Log("NOTICE", v...)
}

func extractTrace(v interface{}) string {
	var trace string
	projectID, err := metadata.ProjectID()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if projectID == "" {
		return ""
	}

	switch t := v.(type) {
	case *http.Request:
		traceHeader := t.Header.Get("X-Cloud-Trace-Context")
		ts := strings.Split(traceHeader, "/")
		if len(ts) > 0 && len(ts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, ts[0])
		}
	default:
		trace = ""
	}

	return trace
}

func (l *Logger) Log(severity string, v ...interface{}) {
	trace := extractTrace(v[0])
	if trace != "" {
		// The first argument was an *http.Request or context object
		// and is not part of the message
		v = v[1:]
	}

	e := Entry{
		Component: l.component,
		Message:   fmt.Sprint(v...),
		Severity:  severity,
		Trace:     trace,
	}

	fmt.Println(e)
}

type Entry struct {
	Message   string `json:"message"`
	Severity  string `json:"severity,omitempty"`
	Trace     string `json:"logging.googleapis.com/trace,omitempty"`
	Component string `json:"component,omitempty"`
}

func (e Entry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	data, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("json.Marshal: %v", err)
	}
	return string(data)
}
