package run

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

// An LogEntry represents a Stackdriver log entry.
type LogEntry struct {
	Message        string                  `json:"message"`
	Severity       string                  `json:"severity,omitempty"`
	Trace          string                  `json:"logging.googleapis.com/trace,omitempty"`
	Component      string                  `json:"component,omitempty"`
	SourceLocation *LogEntrySourceLocation `json:"logging.googleapis.com/sourceLocation,omitempty"`
}

// A LogEntrySourceLocation holds source code location data.
//
// Location data is used to provide additional context when logging
// to Stackdriver.
type LogEntrySourceLocation struct {
	File     string `json:"file,omitempty"`
	Function string `json:"function,omitempty"`
	Line     string `json:"line,omitempty"`
}

// String returns a JSON formatted string expected by Stackdriver.
func (le LogEntry) String() string {
	if le.Severity == "" {
		le.Severity = "INFO"
	}
	data, err := json.Marshal(le)
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

// Info formats using the default formats for its operands and writes to standard output.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to INFO.
func (l *Logger) Info(v ...interface{}) {
	l.Log("INFO", v...)
}

// Error formats using the default formats for its operands and writes to standard output.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to ERROR.
func (l *Logger) Error(v ...interface{}) {
	l.Log("ERROR", v...)
}

// Notice formats using the default formats for its operands and writes to standard output.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to NOTICE.
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

// Log writes logging events with the given severity.
//
// Log formats it's operands using the default format for each value and
// combines the results in to a single log message. If the first value is
// an *http.Request, the X-Cloud-Trace-Context HTTP header will be extracted
// and included in the Stackdriver log entry.
//
// Source file location data will be included in log entires.
//
// Logs are written to stdout in the Stackdriver structured log
// format. See https://cloud.google.com/logging/docs/structured-logging
// for more details.
func (l *Logger) Log(severity string, v ...interface{}) {
	var trace string
	traceID := extractTraceID(v[0])

	if traceID != "" {
		// The first argument was an *http.Request or context object
		// and is not part of the message
		v = v[1:]
		trace = fmt.Sprintf("projects/%s/traces/%s", l.projectID, traceID)
	}

	var sourceLocation *LogEntrySourceLocation
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		sourceLocation = &LogEntrySourceLocation{
			File:     file,
			Line:     strconv.Itoa(line),
			Function: runtime.FuncForPC(pc).Name(),
		}
	}

	e := LogEntry{
		Message:        fmt.Sprint(v...),
		Severity:       severity,
		Trace:          trace,
		SourceLocation: sourceLocation,
	}

	fmt.Println(e)
}
