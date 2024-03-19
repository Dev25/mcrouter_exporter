## 0.3.2 / 2024-03-19

* [CHANGE] Update dependencies

This addresses CVE-2024-24786 which is not exploitable in the exporter, but set off security scanners.

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
