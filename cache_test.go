package run

import (
	"testing"
)

var cacheTests = []struct {
	k string
	v string
}{
	{"test", "http://test.example.com"},
	{"", ""},
}

func TestCache(t *testing.T) {
	data := make(map[string]string)
	c := &cache{data}

	for _, tt := range cacheTests {
		c.Set(tt.k, tt.v)
		v := c.Get(tt.k)
		if v != tt.v {
			t.Errorf("want %s, got %s", tt.v, v)
		}
	}
}
