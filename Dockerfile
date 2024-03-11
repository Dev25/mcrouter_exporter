# Builder
FROM golang:1.22 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:280852156756ea3f39f9e774a30346f2e756244e1f432aea3061c4ac85d90a66
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
