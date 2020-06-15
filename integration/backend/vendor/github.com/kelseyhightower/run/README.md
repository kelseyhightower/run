# run

[![GoDoc](https://godoc.org/github.com/kelseyhightower/run?status.svg)](https://pkg.go.dev/github.com/kelseyhightower/run) ![CloudBuild](https://badger-6bn2iswfgq-ue.a.run.app/build/status?project=hightowerlabs&id=bb0129f8-02c4-490b-b37e-777215fdb7ca)

The run package provides a set of Cloud Run helper functions and does not leverage any third party dependencies.

## Usage

```Go
package main

import (
    "net/http"
    "os"

    "github.com/kelseyhightower/run"
)

func main() {
    // Generates structured logs optimized for Cloud Run.
    run.Notice("Starting helloworld service...")

    // Easy access to secrets stored in Secret Manager.
    secret, err := run.AccessSecret("foo")
    if err != nil {
        run.Fatal(err)
    }

    _ = secret

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Optionally pass in the *http.Request as the first argument
        // to correlate container logs with request logs.
        run.Info(r, "handling http request")

        w.Write([]byte("Hello world!\n"))
    })

    // Start an HTTP server listening on the address defined by the
    // Cloud Run container runtime contract and gracefully shutdown
    // when terminated.
    if err := run.ListenAndServe(nil); err != http.ErrServerClosed {
        run.Fatal(err)
    }
}
```

### Service Authentication

run takes the pain out of [service-to-service authentication](https://cloud.google.com/run/docs/authenticating/service-to-service)

```Go
package main

import (
    "net/http"

    "github.com/kelseyhightower/run"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        request, err := http.NewRequest("GET", "https://example-6bn2iswfgq-uw.a.run.app", nil)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Use the run.Transport to automatically attach ID tokens to outbound requests
        // and optionally expand service names using the Cloud Run API.
        // See https://pkg.go.dev/github.com/kelseyhightower/run?tab=doc#Transport
        client := http.Client{Transport: &run.Transport{EnableServiceNameResolution: false}}

        response, err := client.Do(request)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        defer response.Body.Close()
    })

    if err := run.ListenAndServe(nil); err != http.ErrServerClosed {
        run.Fatal(err)
    }
}
```

## Status

This package is experimental and should not be used or assumed to be stable. Breaking changes are guaranteed to happen.
