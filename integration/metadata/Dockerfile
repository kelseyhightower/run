FROM golang:1.14.3-buster as go-builder
WORKDIR /module
COPY . /module/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo \
      -ldflags '-w -extldflags "-static"' \
      -mod vendor \
      -o run-metadata-integration-tests

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=go-builder /module/run-metadata-integration-tests .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/run-metadata-integration-tests"]
