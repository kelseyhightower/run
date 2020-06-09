package run

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

var loggerTests = []struct {
	message  string
	severity string
}{
	{"info", "INFO"},
	{"error", "ERROR"},
	{"notice", "NOTICE"},
}

func TestLogger(t *testing.T) {
	logger := NewLogger()

	for _, tt := range loggerTests {
		buf := new(bytes.Buffer)
		logger.SetOutput(buf)

		switch tt.severity {
		case "INFO":
			logger.Info(tt.message)
		case "ERROR":
			logger.Error(tt.message)
		case "NOTICE":
			logger.Notice(tt.message)
		}

		var le LogEntry
		if err := json.Unmarshal(buf.Bytes(), &le); err != nil {
			t.Error(err)
		}

		if le.Message != tt.message {
			t.Errorf("log message mismatch, want %s, got %s", tt.message, le.Message)
		}

		if le.Severity != tt.severity {
			t.Errorf("log severity mismatch, want %s, got %s", tt.severity, le.Severity)
		}
	}
}

func TestDefaultLogger(t *testing.T) {
	for _, tt := range loggerTests {
		buf := new(bytes.Buffer)
		SetOutput(buf)

		switch tt.severity {
		case "INFO":
			Info(tt.message)
		case "ERROR":
			Error(tt.message)
		case "NOTICE":
			Notice(tt.message)
		}

		var le LogEntry
		if err := json.Unmarshal(buf.Bytes(), &le); err != nil {
			t.Error(err)
		}

		if le.Message != tt.message {
			t.Errorf("log message mismatch, want %s, got %s", tt.message, le.Message)
		}

		if le.Severity != tt.severity {
			t.Errorf("log severity mismatch, want %s, got %s", tt.severity, le.Severity)
		}
	}
}

func TestLoggerWithTraceID(t *testing.T) {
	var (
		traceID         = "27abb75176a19ccf353146b192ef419f"
		formatedTraceID = fmt.Sprintf("projects/%s/traces/%s", gcptest.ProjectID, traceID)
		message         = "message"
		severity        = "INFO"
	)

	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	buf := new(bytes.Buffer)
	SetOutput(buf)

	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Error(err)
	}

	r.Header.Set("X-Cloud-Trace-Context", traceID)

	Info(r, message)

	var le LogEntry
	if err := json.Unmarshal(buf.Bytes(), &le); err != nil {
		t.Error(err)
	}

	if le.Trace != formatedTraceID {
		t.Errorf("log traceID mismatch, want %s, got %s", formatedTraceID, le.Trace)
	}

	if le.Message != message {
		t.Errorf("log message mismatch, want %s, got %s", message, le.Message)
	}

	if le.Severity != severity {
		t.Errorf("log severity mismatch, want %s, got %s", severity, le.Severity)
	}
}
