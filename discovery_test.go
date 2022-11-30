package run

var formatEndpointIDTests = []struct {
	name string
	ip   string
	want string
}{
	{"example", "10.0.0.1", "example-10-0-0-1"},
	{"example-service", "10.0.0.1", "example-service-10-0-0-1"},
}
