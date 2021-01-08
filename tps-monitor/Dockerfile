FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /tps-monitor
RUN cd /tps-monitor && make

# Pull tpsmonitor into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /tps-monitor/build/bin/tpsmonitor /usr/local/bin/

EXPOSE 7575
ENTRYPOINT ["tpsmonitor"]