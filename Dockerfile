FROM golang:alpine AS builder
COPY . /go/src/github.com/dev25/mcrouter_exporter
WORKDIR /go/src/github.com/dev25/mcrouter_exporter
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /app

FROM quay.io/prometheus/busybox:latest
COPY --from=builder /app /usr/local/bin/mcrouter_exporter
EXPOSE 9151
CMD ["/usr/local/bin/mcrouter_exporter"]
