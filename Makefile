PKG := github.com/dev25/mcrouter_exporter
IMAGE := quay.io/dev25/mcrouter-exporter
OUT := mcrouter_exporter

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

setup:
	go get -v -u ./...

fmt:
	@go fmt

vet:
	@go vet ${PKG_LIST}

test: fmt vet
	go test

build:
	go build -i -v -o ${OUT} -ldflags=$(FLAGS)

run:
	./$(OUT)

clean:
	-@rm -f ${OUT} ${OUT}_docker

docker:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o ${OUT}_docker .
	docker build -t $(IMAGE) .
	rm -f ${OUT}_docker

.PHONY: all build test docker vet clean
