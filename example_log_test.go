package run_test

import (
	"log"

	"github.com/kelseyhightower/run"
)

func ExampleLogger() {
	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting example service...")
}
