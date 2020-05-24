package run

import "os"

// Port returns the port your HTTP server should listen on.
func Port() string {
	return os.Getenv("PORT")
}

// Revision returns the name of the Cloud Run revision being run.
func Revision() string {
	return os.Getenv("K_REVISION")
}

// Configuration returns the name of the Cloud Run configuration being run.
func Configuration() string {
	return os.Getenv("K_CONFIGURATION")
}

// ServiceName returns the name of the Cloud Run service being run.
func ServiceName() string {
	return os.Getenv("K_SERVICE")
}
