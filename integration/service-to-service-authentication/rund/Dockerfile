FROM golang:1.14-buster as go-builder
ENV GO111MODULE=on
WORKDIR /module
COPY . /module/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo \
      -ldflags '-w -extldflags "-static"' \
      -o rund

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=go-build /rund .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/rund"]
