# Builder
FROM golang:1.14.4 as builder
WORKDIR /workspace
COPY . /workspace
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o mcrouter_exporter main.go

# Use distroless as final image
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
