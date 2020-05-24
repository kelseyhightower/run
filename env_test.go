package run

import (
	"os"
	"testing"
)

var envTests = []struct {
	n    string
	e    string
	want string
	f    func() string
}{
	{"Configuration", "K_CONFIGURATION", "test", Configuration},
	{"Revision", "K_REVISION", "test-0001-vob", Revision},
	{"Service", "K_SERVICE", "test", ServiceName},
	{"Port", "PORT", "8080", Port},
}

func TestEnv(t *testing.T) {
	for _, tt := range envTests {
		t.Run(tt.n, func(t *testing.T) {
			if err := os.Setenv(tt.e, tt.want); err != nil {
				t.Error(err)
			}

			got := tt.f()
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

