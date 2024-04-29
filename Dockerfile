# Builder
FROM golang:1.22 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:0bb1e72361cf6aa3f66af29360da60220b9a8fc8b063dfa634d16e68c26c94f0
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
