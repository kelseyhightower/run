package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown waits for the SIGKILL, SIGINT, or SIGTERM signals and shutdowns the process.
func WaitForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	s := <-signalChan

	Notice(fmt.Sprintf("Received shutdown signal: %v; shutdown complete.", s.String()))
}
