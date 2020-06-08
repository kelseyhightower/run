package run

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
)

var DefaultLogger = &Logger{out: os.Stdout}

// An LogEntry represents a Stackdriver log entry.
type LogEntry struct {
	Message        string                  `json:"message"`
	Severity       string                  `json:"severity,omitempty"`
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
	mu  sync.Mutex
	buf []byte
	out io.Writer
}

// NewLogger creates a new Logger.
func NewLogger() *Logger {
	return &Logger{out: os.Stdout}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// Info formats using the default formats for its operands.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to INFO.
func (l *Logger) Info(v ...interface{}) {
	l.Log("INFO", v...)
}

// Error formats using the default formats for its operands.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to ERROR.
func (l *Logger) Error(v ...interface{}) {
	l.Log("ERROR", v...)
}

// Notice formats using the default formats for its operands.
//
// Logs are written in the Stackdriver structured logging format with the severity level
// set to NOTICE.
func (l *Logger) Notice(v ...interface{}) {
	l.Log("NOTICE", v...)
}

// Log writes logging events with the given severity.
//
// Log formats it's operands using the default format for each value and
// combines the results in to a single log message.
//
// Source file location data will be included in log entires.
//
// Logs are written to stdout in the Stackdriver structured log
// format. See https://cloud.google.com/logging/docs/structured-logging
// for more details.
func (l *Logger) Log(severity string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

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
		SourceLocation: sourceLocation,
	}

	l.buf = l.buf[:0]
	l.buf = append(l.buf, e.String()...)
	l.out.Write(l.buf)
}
