package run

import (
	"bytes"
	"encoding/json"
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
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	logger, err := NewLogger()
	if err != nil {
		t.Error(err)
	}

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
