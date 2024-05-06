# Builder
FROM golang:1.22 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:2fb55308ef768a0ca0851f294d7f5b582579dba6522d1d2162e2d5f33b876e97
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
