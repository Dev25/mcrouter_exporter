Mcrouter Prometheus Exporter
===
[![Docker Repository on Quay](https://quay.io/repository/dev25/mcrouter_exporter/status "Docker Repository on Quay")](https://quay.io/repository/dev25/mcrouter_exporter)
[![CircleCI](https://circleci.com/gh/Dev25/mcrouter_exporter.svg?style=svg)](https://circleci.com/gh/Dev25/mcrouter_exporter)

Prometheus exporter for [mcrouter](https://github.com/facebook/mcrouter), a popular memcache router developed by Facebook

Building
---

By default the `mcrouter_exporter` will scrape mcrouter metrics on `localhost:5000` and expose the metrics for Prometheus consumption on `0.0.0.0:9442`. This can be configured using either `-mcrouter.address` or `web.listen-address` flags.

```
go get -v -u github.com/Dev25/mcrouter_exporter
cd $GOPATH/src/github.com/Dev25/mcrouter_exporter
make
./mcrouter_exporter
```

Docker Images
----
Docker images have been created for both mcrouter and mcrouter_exporter, these can be found at:

- [quay.io/dev25/mcrouter:v37](https://quay.io/repository/dev25/mcrouter?tab=tags)
- [quay.io/dev25/mcrouter_exporter](https://quay.io/repository/dev25/mcrouter_exporter?tab=tags)


Collectors
----
The exporter collects a number of statistics from mcrouter:

```
# HELP mcrouter_asynclog_requests Number of failed deletes written to spool file.
# TYPE mcrouter_asynclog_requests counter
# HELP mcrouter_asynclog_requests_rate Number of requests that were attempted to be spooled to disk.
# TYPE mcrouter_asynclog_requests_rate gauge
# HELP mcrouter_asynclog_spool_success_rate Number of requests that were spooled successfully.
# TYPE mcrouter_asynclog_spool_success_rate gauge
# HELP mcrouter_clients Number of connected clients.
# TYPE mcrouter_clients counter
# HELP mcrouter_command_count Total number of received requests drilled down by operation.
# TYPE mcrouter_command_count counter
# HELP mcrouter_command_out Average number of sent normal (non-shadow, non-failover) requests per second drilled
# TYPE mcrouter_command_out counter
# HELP mcrouter_command_out_all Total number of sent requests per second (failover + shadow + normal)
# TYPE mcrouter_command_out_all counter
# HELP mcrouter_commandargs Command args used.
# TYPE mcrouter_commandargs gauge
# HELP mcrouter_commands Average number of received requests per second drilled down by operation.
# TYPE mcrouter_commands gauge
# HELP mcrouter_config_failures How long ago (in seconds) mcrouter has reconfigured.
# TYPE mcrouter_config_failures counter
# HELP mcrouter_config_last_attempt How long ago (in seconds) mcrouter has reconfigured.
# TYPE mcrouter_config_last_attempt gauge
# HELP mcrouter_config_last_success How long ago (in seconds) mcrouter has reconfigured.
# TYPE mcrouter_config_last_success gauge
# HELP mcrouter_cpu_seconds_total Number of seconds mcrouter spent on CPU.
# TYPE mcrouter_cpu_seconds_total counter
# HELP mcrouter_dev_null_requests Number of requests sent to DevNullRoute.
# TYPE mcrouter_dev_null_requests counter
# HELP mcrouter_duration_us Average time of processing a request (i.e. receiving request and sending a reply).
# TYPE mcrouter_duration_us gauge
# HELP mcrouter_fibers_allocated Number of fibers (lightweight threads) created by mcrouter.
# TYPE mcrouter_fibers_allocated gauge
# HELP mcrouter_proxy_reqs_processing Requests mcrouter started routing but didn't receive a reply yet.
# TYPE mcrouter_proxy_reqs_processing gauge
# HELP mcrouter_proxy_reqs_waiting Requests queued up and not routed yet.
# TYPE mcrouter_proxy_reqs_waiting gauge
# HELP mcrouter_request TODO.
# TYPE mcrouter_request gauge
# HELP mcrouter_request_count TODO
# TYPE mcrouter_request_count counter
# HELP mcrouter_resident_memory_bytes Number of bytes of resident memory.
# TYPE mcrouter_resident_memory_bytes counter
# HELP mcrouter_result_all Average number of replies per second received for requests drilled down by reply
# TYPE mcrouter_result_all gauge
# HELP mcrouter_result_all_count TODO.
# TYPE mcrouter_result_all_count counter
# HELP mcrouter_result_count Total number of replies received drilled down by reply result
# TYPE mcrouter_result_count counter
# HELP mcrouter_results Average number of replies per second received for normal requests drilled down by reply
# TYPE mcrouter_results gauge
# HELP mcrouter_servers Number of connected memcached servers.
# TYPE mcrouter_servers gauge
# HELP mcrouter_start_time_seconds The timestamp of mcrouter daemon start.
# TYPE mcrouter_start_time_seconds counter
# HELP mcrouter_up Could the mcrouter server be reached.
# TYPE mcrouter_up gauge
# HELP mcrouter_version Version of mcrouter binary.
# TYPE mcrouter_version gauge
# HELP mcrouter_virtual_memory_bytes Number of bytes of virtual memory.
# TYPE mcrouter_virtual_memory_bytes counter
```

Optional metrics available when setting the mcrouter.server_metrics argument:

```
# HELP mcrouter_server_duration_us Average time of processing a request per-server (i.e. receiving request and sending a reply).
# TYPE mcrouter_server_duration_us gauge
# HELP mcrouter_server_memcached_connect_timeout_count Number of memcached connect timeouts (per-server metric).
# TYPE mcrouter_server_memcached_connect_timeout_count counter
# HELP mcrouter_server_memcached_deleted_count Number of memcached DELETED replies (per-server metric).
# TYPE mcrouter_server_memcached_deleted_count counter
# HELP mcrouter_server_memcached_exists_count Number of memcached EXISTS replies (per-server metric).
# TYPE mcrouter_server_memcached_exists_count counter
# HELP mcrouter_server_memcached_found_count Number of memcached FOUND replies (per-server metric).
# TYPE mcrouter_server_memcached_found_count counter
# HELP mcrouter_server_memcached_not_found_count Number of memcached NOT_FOUND replies (per-server metric).
# TYPE mcrouter_server_memcached_not_found_count counter
# HELP mcrouter_server_memcached_not_stored_count Number of memcached NOT_STORED replies (per-server metric).
# TYPE mcrouter_server_memcached_not_stored_count counter
# HELP mcrouter_server_memcached_remote_error_count Number of memcached remote errors (per-server metric).
# TYPE mcrouter_server_memcached_remote_error_count counter
# HELP mcrouter_server_memcached_stored_count Number of memcached STORED replies (per-server metric).
# TYPE mcrouter_server_memcached_stored_count counter
# HELP mcrouter_server_memcached_timeout_count Number of memcached timeouts (per-server metric).
# TYPE mcrouter_server_memcached_timeout_count counter
# HELP mcrouter_server_memcached_tko Number of times memcached has been marked as TKO (per-server metric).
# TYPE mcrouter_server_memcached_tko counter
# HELP mcrouter_server_memcached_touched_count Number of memcached TOUCHED replies (per-server metric).
# TYPE mcrouter_server_memcached_touched_count counter
# HELP mcrouter_server_proxy_reqs_processing Requests mcrouter started routing but didn't receive a reply yet (per-server metric)
# TYPE mcrouter_server_proxy_reqs_processing gauge
# HELP mcrouter_server_proxy_reqs_retrans_ratio Requests mcrouter started but that required retransmission.
# TYPE mcrouter_server_proxy_reqs_retrans_ratio gauge
# HELP mcrouter_server_proxy_reqs_waiting Requests queued up and not routed yet (per-server metric)
# TYPE mcrouter_server_proxy_reqs_waiting gauge
```
