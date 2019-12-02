package run

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// An Entry represents a Stackdriver log entry.
type Entry struct {
	Message   string `json:"message"`
	Severity  string `json:"severity,omitempty"`
	Trace     string `json:"logging.googleapis.com/trace,omitempty"`
	Component string `json:"component,omitempty"`
}

// String returns a JSON formatted string expected by Stackdriver.
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

// A Logger represents an active logging object that generates JSON formatted
// log entries to standard out. Logs are formatted as expected by Cloud Run's
// Stackdriver integration.
type Logger struct {
	projectID string
}

// NewLogger creates a new Logger.
func NewLogger() (*Logger, error) {
	projectID, err := ProjectID()
	if err != nil {
		return nil, err
	}

	return &Logger{projectID: projectID}, nil
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

func extractTraceID(v interface{}) string {
	var trace string

	switch t := v.(type) {
	case *http.Request:
		traceHeader := t.Header.Get("X-Cloud-Trace-Context")
		ts := strings.Split(traceHeader, "/")
		if len(ts) > 0 && len(ts[0]) > 0 {
			trace = ts[0]
		}
	default:
		trace = ""
	}

	return trace
}

func (l *Logger) Log(severity string, v ...interface{}) {
	var trace string
	traceID := extractTraceID(v[0])

	if traceID != "" {
		// The first argument was an *http.Request or context object
		// and is not part of the message
		v = v[1:]
		trace = fmt.Sprintf("projects/%s/traces/%s", l.projectID, traceID)
	}

	e := Entry{
		Message:  fmt.Sprint(v...),
		Severity: severity,
		Trace:    trace,
	}

	fmt.Println(e)
}
