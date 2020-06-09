package run

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var dl = &Logger{out: os.Stdout}

// An LogEntry represents a Stackdriver log entry.
type LogEntry struct {
	Message        string                  `json:"message"`
	Severity       string                  `json:"severity,omitempty"`
	Component      string                  `json:"component,omitempty"`
	SourceLocation *LogEntrySourceLocation `json:"logging.googleapis.com/sourceLocation,omitempty"`
	Trace          string                  `json:"logging.googleapis.com/trace,omitempty"`
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
	mu  sync.Mutex
	buf []byte
	out io.Writer
}

// Error calls Log on the default logger with severity set to ERROR.
//
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	dl.Log("ERROR", v...)
}

// Fatal calls Log on the default logger with severity set to ERROR
// followed by a call to os.Exit(1).
//
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	dl.Log("ERROR", v...)
	os.Exit(1)
}

// Log writes logging events with the given severity.
//
// The string s contains the text to log.
//
// Source file location data will be included in log entires.
//
// Logs are written to stdout in the Stackdriver structured log
// format. See https://cloud.google.com/logging/docs/structured-logging
// for more details.
func Log(severity, s string) {
	dl.Log(severity, s)
}

// SetOutput sets the output destination for the default logger.
func SetOutput(w io.Writer) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.out = w
}

// Info calls Log on the default logger with severity set to INFO.
//
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	dl.Log("INFO", v...)
}

// Notice calls Log on the default logger with severity set to NOTICE.
//
// Arguments are handled in the manner of fmt.Print.
func Notice(v ...interface{}) {
	dl.Log("NOTICE", v...)
}

// NewLogger creates a new Logger.
func NewLogger() *Logger {
	return &Logger{out: os.Stdout}
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// Info calls l.Log with severity set to INFO.
//
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.Log("INFO", v...)
}

// Error calls l.Log with severity set to ERROR.
//
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.Log("ERROR", v...)
}

// Fatal calls l.Log with severity set to ERROR followed by
// a call to os.Exit(1).
//
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	l.Log("ERROR", v...)
	os.Exit(1)
}

// Notice calls l.Log with severity set to NOTICE.
//
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Notice(v ...interface{}) {
	l.Log("NOTICE", v...)
}

// Log writes logging events with the given severity.
//
// If the first value is an *http.Request, the X-Cloud-Trace-Context
// HTTP header will be extracted and included in the Stackdriver log
// entry.
//
// Source file location data will be included in log entires.
//
// Logs are written to stdout in the Stackdriver structured log
// format. See https://cloud.google.com/logging/docs/structured-logging
// for more details.
func (l *Logger) Log(severity string, v ...interface{}) {
	var traceID string

	tid := extractTraceID(v[0])
	if tid != "" {
		// The first argument was an *http.Request or context object
		// and is not part of the message
		v = v[1:]

		pid, err := ProjectID()
		if err != nil {
			e := &LogEntry{
				Message:  fmt.Sprintf("unable to append trace to log, missing project id: %v", err.Error()),
				Severity: "ERROR",
			}
			l.write(e)
		}

		if pid == "" {
			e := &LogEntry{
				Message:  fmt.Sprint("unable to append trace to log, project id is empty"),
				Severity: "ERROR",
			}
			l.write(e)
		} else {
			traceID = fmt.Sprintf("projects/%s/traces/%s", pid, tid)
		}
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

	e := &LogEntry{
		Message:        fmt.Sprint(v...),
		Severity:       severity,
		SourceLocation: sourceLocation,
		Trace:          traceID,
	}

	l.write(e)
}

func (l *Logger) write(e *LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := e.String()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	l.out.Write(l.buf)
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
