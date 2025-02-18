## 0.5.0 / 2025-02-18

* [CHANGE] Update to Go 1.24
* [CHANGE] Update distroless/base-debian11
* [ENHANCEMENT] Export fiber pool size stat

## 0.4.0 / 2024-03-27

* [CHANGE] Update distroless/base-debian11
* [CHANGE] Update to Go 1.22
* [CHANGE] Update dependencies
* [CHANGE] Add support for multi arch/arm64 images hosted on ghcr.io
* [CHANGE] Github hosted image path to ghcr.io/dev25/mcrouter_exporter

## 0.3.1 / 2023-12-23

* [FIXED] CI Releasing to quay.io

## 0.3.0 / 2023-12-23

* [CHANGE] Update to Go 1.21
* [ENHANCEMENT] Add new rate asynclog metrics #38

## 0.2.0 / 2023-02-21

* [CHANGE] Update to Go 1.20
* [CHANGE] Publish images to ghcr.io/dev25/mcrouter_exporter/mcrouter_exporter
* [CHANGE] Move to distroless base image
* [ENHANCEMENT] Support for client connections in mcrouter v39+
* [ENHANCEMENT] Add TKO reply per-server metrics
* [FIXED] Parse server memcached notfound and notstored metrics correctly

## 0.1.0 / 2019-11-23

* [CHANGE] Update prometheus/client_golang to 1.1.0
* [ENHANCEMENT] Build branch based and tagged docker images in CircleCI
