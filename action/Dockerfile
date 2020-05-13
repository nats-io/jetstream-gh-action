FROM alpine:latest

RUN apk add --update ca-certificates

COPY action /

ENTRYPOINT ["/action"]

