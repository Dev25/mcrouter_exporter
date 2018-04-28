# Mcrouter Prometheus Exporter
---

Prometheus exporter for [mcrouter](https://github.com/facebook/mcrouter), a popular memcache router developed by Facebook

Building
---

By default the `mcrouter_exporter` will scrape mcrouter metrics on `localhost:5000` and expose the metrics for Prometheus consumption on `0.0.0.0:9151`. This can be configured using either `-mcrouter.address` or `web.listen-address` flags.

```
go get -v -u github.com/Dev25/mcrouter_exporter
cd $GOPATH/src/github.com/Dev25/mcrouter_exporter
make
./mcrouter_exporter
```

A Dockerfile has also been provided to build Docker images.
```
make docker IMAGE=mcrouter_exporter
docker run --rm mcrouter_exporter
```

Docker Images
----
Docker images have been created for both mcrouter and mcrouter_exporter, these can be found at:

- `devan2502/mcrouter:v37`
- `devan2502/mcrouter_exporter`
