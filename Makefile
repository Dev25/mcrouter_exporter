PKG := github.com/dev25/mcrouter_exporter
OUT := mcrouter_exporter

GOLANGCI_VERSION ?= 1.27.0

# Build info
REVISION := $(shell git describe --always --long --dirty)
VERSION := $(shell git tag --points-at HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)

FLAGS := "-X github.com/prometheus/common/version.Version=${VERSION} \
	-X github.com/prometheus/common/version.Branch=${BRANCH} \
	-X github.com/prometheus/common/version.Revision=${REVISION} \
	-X github.com/prometheus/common/version.BuildDate=${BUILD_DATE}"

all: build


bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	bin/golangci-lint run --fix


fmt:
	@go fmt

vet:
	@go vet ${PKG_LIST}

test: fmt vet
	go test

build:
	go build -i -v -o ${OUT} -ldflags=$(FLAGS)

docker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -a -o ${OUT} -ldflags=$(FLAGS)

clean:
	-@rm -f ${OUT}

.PHONY: all setup fmt test build clean
