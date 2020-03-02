FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY runc .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/runc"]
