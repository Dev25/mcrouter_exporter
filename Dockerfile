# Builder
FROM golang:1.22 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:9bc3117a99c731a41200a28774405125cb6fbda1819f4a1af88bd3bfad5dcf32
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
