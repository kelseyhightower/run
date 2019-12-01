package run

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	component string
	out       io.Writer // destination for output
}

func NewLogger(component string) *Logger {
	return &Logger{out: os.Stdout, component: component}
}

func (l *Logger) LogError(v ...interface{}) {
	l.Log("ERROR", v...)
}

func (l *Logger) LogNotice(v ...interface{}) {
	l.Log("NOTICE", v...)
}

func (l *Logger) Log(severity string, v ...interface{}) {
	e := Entry{
		Message:   fmt.Sprint(v...),
		Severity:  severity,
		Component: l.component,
	}
	log.Println(e)
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
		log.Printf("json.Marshal: %v", err)
	}
	return string(data)
}
