# Builder
FROM golang:1.20 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:7b9dc0fa2731bfddc1a94c84994bd2ef87b2d89721596331fc63c5403b8c3f64
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
