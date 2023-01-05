# Builder
FROM golang:1.17 as builder
WORKDIR /workspace
COPY . /workspace
RUN make build-docker

# Use distroless as final image
FROM gcr.io/distroless/base-debian11@sha256:e5853c0285c4c07ab5724d0b582c9b168f6c8dfa330627d22f814d98d77c5b85
WORKDIR /
COPY --from=builder /workspace/mcrouter_exporter .
ENTRYPOINT ["/mcrouter_exporter"]
