# Builder
FROM golang:1.22 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:6c1e34e2f084fe6df17b8bceb1416f1e11af0fcdb1cef11ee4ac8ae127cb507c
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
