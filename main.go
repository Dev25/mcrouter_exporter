package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
)

const (
	namespace = "mcrouter"
)

type Exporter struct {
	server         string
	timeout        time.Duration
	server_stats   bool
	admin_requests bool
	logger         log.Logger

	up                            *prometheus.Desc
	startTime                     *prometheus.Desc
	version                       *prometheus.Desc
	commandArgs                   *prometheus.Desc
	commands                      *prometheus.Desc
	commandCount                  *prometheus.Desc
	commandOut                    *prometheus.Desc
	commandOutFailover            *prometheus.Desc
	commandOutShadow              *prometheus.Desc
	commandOutAll                 *prometheus.Desc
	commandOutCount               *prometheus.Desc
	configFailures                *prometheus.Desc
	configLastAttempt             *prometheus.Desc
	configLastSuccess             *prometheus.Desc
	devNullRequests               *prometheus.Desc
	duration                      *prometheus.Desc
	fibersAllocated               *prometheus.Desc
	fibersPoolSize                *prometheus.Desc
	proxyReqsProcessing           *prometheus.Desc
	proxyReqsWaiting              *prometheus.Desc
	requests                      *prometheus.Desc
	requestCount                  *prometheus.Desc
	results                       *prometheus.Desc
	resultCount                   *prometheus.Desc
	resultFailover                *prometheus.Desc
	resultShadow                  *prometheus.Desc
	resultAll                     *prometheus.Desc
	resultAllCount                *prometheus.Desc
	clients                       *prometheus.Desc
	numClientConnections          *prometheus.Desc
	servers                       *prometheus.Desc
	cpuSeconds                    *prometheus.Desc
	residentMemory                *prometheus.Desc
	virtualMemory                 *prometheus.Desc
	asynclogRequests              *prometheus.Desc
	asynclogRequestsRate          *prometheus.Desc
	asynclogSpoolSuccessRate      *prometheus.Desc
	serverDuration                *prometheus.Desc
	serverProxyReqsProcessing     *prometheus.Desc
	serverProxyInflightReqs       *prometheus.Desc
	serverProxyRetransRatio       *prometheus.Desc
	serverMemcachedStored         *prometheus.Desc
	serverMemcachedNotStored      *prometheus.Desc
	serverMemcachedFound          *prometheus.Desc
	serverMemcachedNotFound       *prometheus.Desc
	serverMemcachedDeleted        *prometheus.Desc
	serverMemcachedTouched        *prometheus.Desc
	serverMemcachedExists         *prometheus.Desc
	serverMemcachedRemoteError    *prometheus.Desc
	serverMemcachedConnectTimeout *prometheus.Desc
	serverMemcachedTimeout        *prometheus.Desc
	serverMemcachedSoftTKO        *prometheus.Desc
	serverMemcachedHardTKO        *prometheus.Desc
	adminRequestVersion           *prometheus.Desc
	adminRequestConfigAge         *prometheus.Desc
	adminRequestConfigFile        *prometheus.Desc
	adminRequestHostId            *prometheus.Desc
	adminRequestConfigMD5Digest   *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter(server string, timeout time.Duration, server_stats bool, admin_requests bool, logger log.Logger) *Exporter {
	return &Exporter{
		server:         server,
		timeout:        timeout,
		server_stats:   server_stats,
		admin_requests: admin_requests,
		logger:         logger,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the mcrouter server be reached.",
			nil,
			nil,
		),
		startTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_time_seconds"),
			"UNIX timestamp of mcrouter startup time.",
			nil,
			nil,
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "version"),
			"Version of mcrouter binary.",
			[]string{"version"},
			nil,
		),
		commandArgs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "commandargs"),
			"Command line arguments used to start mcrouter.",
			[]string{"commandargs"},
			nil,
		),
		commands: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "commands"),
			"Average number of received requests per second drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		commandCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_count"),
			"Total number of received requests drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		commandOut: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_out"),
			"Average number of sent normal (non-shadow, non-failover) requests per second drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		commandOutFailover: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_out_failover"),
			"Average number of sent failover requests per second drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		commandOutShadow: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_out_shadow"),
			"Number of sent shadow requests per second drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		commandOutAll: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_out_all"),
			"Total number of sent requests per second (failover + shadow + normal)",
			[]string{"cmd"},
			nil,
		),
		commandOutCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "command_out_count"),
			"Total number of sent requests per second drilled down by operation.",
			[]string{"cmd"},
			nil,
		),
		configFailures: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "config_failures"),
			"How many times mcrouter failed to reconfigure (if > 0 and growing, check the config is valid).",
			nil,
			nil,
		),
		configLastAttempt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "config_last_attempt"),
			"UNIX timestamp of last time mcrouter tried to reconfigure.",
			nil,
			nil,
		),
		configLastSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "config_last_success"),
			"UNIX timestamp of last time mcrouter reconfigured successfully.",
			nil,
			nil,
		),
		devNullRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "dev_null_requests"),
			"Number of requests sent to DevNullRoute.",
			nil,
			nil,
		),
		duration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "duration_us"),
			"Average time of processing a request (i.e. receiving request and sending a reply).",
			nil,
			nil,
		),
		fibersAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "fibers_allocated"),
			"Number of fibers (lightweight threads) created by mcrouter.",
			nil,
			nil,
		),
		fibersPoolSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "fibers_pool_size"),
			"Number of fibers (lightweight threads) created by mcrouter that are currently in the free pool.",
			nil,
			nil,
		),
		proxyReqsProcessing: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "proxy_reqs_processing"),
			"Requests mcrouter started routing but didn't receive a reply yet.",
			nil,
			nil,
		),
		proxyReqsWaiting: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "proxy_reqs_waiting"),
			"Requests queued up and not routed yet.",
			nil,
			nil,
		),
		clients: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "clients"),
			"Number of connected clients (prior to version 39).",
			nil,
			nil,
		),
		numClientConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "num_client_connections"),
			"Number of connected clients (version 39 and after).",
			nil,
			nil,
		),
		servers: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "servers"),
			"Number of connected memcached servers.",
			[]string{"state"},
			nil,
		),
		requests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "request"),
			"TODO.",
			[]string{"type"},
			nil,
		),
		requestCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "request_count"),
			"TODO",
			[]string{"type"},
			nil,
		),
		// Result
		results: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "results"),
			"Average number of replies per second received for normal requests drilled down by reply result.",
			[]string{"reply"},
			nil,
		),
		resultCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "result_count"),
			"Total number of replies received drilled down by reply result",
			[]string{"reply"},
			nil,
		),
		resultFailover: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "result_failover"),
			"Average number of replies per second received for failover requests drilled down by result.",
			[]string{"reply"},
			nil,
		),
		resultShadow: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "result_shadow"),
			"Average number of replies per second received for shadow requests drilled down by result.",
			[]string{"reply"},
			nil,
		),
		resultAll: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "result_all"),
			"Average number of replies per second received for requests drilled down by reply result.",
			[]string{"reply"},
			nil,
		),
		resultAllCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "result_all_count"),
			"TODO.",
			[]string{"reply"},
			nil,
		),
		cpuSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "cpu_seconds_total"),
			"Number of seconds mcrouter spent on CPU.",
			nil,
			nil,
		),
		residentMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "resident_memory_bytes"),
			"Number of bytes of resident memory.",
			nil,
			nil,
		),
		virtualMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "virtual_memory_bytes"),
			"Number of bytes of virtual memory.",
			nil,
			nil,
		),
		asynclogRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "asynclog_requests"),
			"Number of failed deletes written to spool file.",
			nil,
			nil,
		),
		asynclogRequestsRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "asynclog_requests_rate"),
			"Number of requests that were attempted to be spooled to disk.",
			nil,
			nil,
		),
		asynclogSpoolSuccessRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "asynclog_spool_success_rate"),
			"Number of requests that were spooled successfully.",
			nil,
			nil,
		),
		serverDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_duration_us"),
			"Average time of processing a request per-server (i.e. receiving request and sending a reply).",
			[]string{"server"},
			nil,
		),
		serverProxyReqsProcessing: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_proxy_reqs_processing"),
			"Requests mcrouter started routing but didn't receive a reply yet (per-server metric)",
			[]string{"server"},
			nil,
		),
		serverProxyInflightReqs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_proxy_reqs_waiting"),
			"Requests queued up and not routed yet (per-server metric)",
			[]string{"server"},
			nil,
		),
		serverProxyRetransRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_proxy_reqs_retrans_ratio"),
			"Requests mcrouter started but that required retransmission.",
			[]string{"server"},
			nil,
		),
		serverMemcachedStored: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_stored_count"),
			"Number of memcached STORED replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedNotStored: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_not_stored_count"),
			"Number of memcached NOT_STORED replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedFound: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_found_count"),
			"Number of memcached FOUND replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedNotFound: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_not_found_count"),
			"Number of memcached NOT_FOUND replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedDeleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_deleted_count"),
			"Number of memcached DELETED replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedTouched: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_touched_count"),
			"Number of memcached TOUCHED replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedExists: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_exists_count"),
			"Number of memcached EXISTS replies (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedRemoteError: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_remote_error_count"),
			"Number of memcached remote errors (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedConnectTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_connect_timeout_count"),
			"Number of memcached connect timeouts (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_timeout_count"),
			"Number of memcached timeouts (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedSoftTKO: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_soft_tko"),
			"Whether or not memcached has been marked as Soft TKO (per-server metric).",
			[]string{"server"},
			nil,
		),
		serverMemcachedHardTKO: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_memcached_hard_tko"),
			"Whether or not memcached has been marked as Hard TKO (per-server metric).",
			[]string{"server"},
			nil,
		),
		adminRequestVersion: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "admin_request_version"),
			"Version string of the build (same string as returned by version).",
			[]string{"version"},
			nil,
		),
		adminRequestConfigAge: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "admin_request_config_age"),
			"How long, in seconds, since last config reload.",
			nil,
			nil,
		),
		adminRequestConfigFile: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "admin_request_config_file"),
			"Config file location (error if configured from string).",
			[]string{"file"},
			nil,
		),
		adminRequestHostId: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "admin_request_host_id"),
			"Hostid of this mcrouter instance, an unsigned 32-bit integer in decimal.",
			[]string{"host_id"},
			nil,
		),
		adminRequestConfigMD5Digest: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "admin_request_config_md5_digest"),
			"Current config's md5 hash. Note that this only specifies the main config file's md5 and ignores any additional tracked files.",
			[]string{"hash"},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the mcrouter exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.startTime
	ch <- e.version
	ch <- e.commands
	ch <- e.commandCount
	ch <- e.commandOut
	ch <- e.commandOutFailover
	ch <- e.commandOutShadow
	ch <- e.commandOutAll
	ch <- e.commandOutCount
	ch <- e.configFailures
	ch <- e.configLastAttempt
	ch <- e.configLastSuccess
	ch <- e.devNullRequests
	ch <- e.duration
	ch <- e.fibersAllocated
	ch <- e.proxyReqsProcessing
	ch <- e.proxyReqsWaiting
	ch <- e.requests
	ch <- e.requestCount
	ch <- e.results
	ch <- e.resultCount
	ch <- e.resultFailover
	ch <- e.resultShadow
	ch <- e.resultAll
	ch <- e.resultAllCount
	ch <- e.clients
	ch <- e.numClientConnections
	ch <- e.servers
	ch <- e.cpuSeconds
	ch <- e.residentMemory
	ch <- e.virtualMemory
	ch <- e.asynclogRequests
	ch <- e.asynclogRequestsRate
	ch <- e.asynclogSpoolSuccessRate

	if e.server_stats {
		ch <- e.serverDuration
		ch <- e.serverProxyReqsProcessing
		ch <- e.serverProxyInflightReqs
		ch <- e.serverProxyRetransRatio
		ch <- e.serverMemcachedStored
		ch <- e.serverMemcachedNotStored
		ch <- e.serverMemcachedFound
		ch <- e.serverMemcachedNotFound
		ch <- e.serverMemcachedDeleted
		ch <- e.serverMemcachedTouched
		ch <- e.serverMemcachedExists
		ch <- e.serverMemcachedRemoteError
		ch <- e.serverMemcachedConnectTimeout
		ch <- e.serverMemcachedTimeout
		ch <- e.serverMemcachedSoftTKO
		ch <- e.serverMemcachedHardTKO
	}

	if e.admin_requests {
		ch <- e.adminRequestVersion
		ch <- e.adminRequestConfigAge
		ch <- e.adminRequestConfigFile
		ch <- e.adminRequestHostId
		ch <- e.adminRequestConfigMD5Digest
	}
}

// Collect fetches the statistics from the configured mcrouter server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	network := "tcp"
	if strings.Contains(e.server, "/") {
		network = "unix"
	}

	conn, err := net.DialTimeout(network, e.server, e.timeout)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		level.Error(e.logger).Log("msg", "Failed to collect stats from mcrouter", "err", err)
		return
	}
	defer conn.Close()

	s, err := getStats(conn)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		level.Error(e.logger).Log("msg", "Failed to collect stats from mcrouter", "err", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	// Parse basic stats
	ch <- prometheus.MustNewConstMetric(e.startTime, prometheus.CounterValue, e.parse(s, "start_time"))
	ch <- prometheus.MustNewConstMetric(e.version, prometheus.GaugeValue, 1, s["version"])
	ch <- prometheus.MustNewConstMetric(e.commandArgs, prometheus.GaugeValue, 1, s["commandargs"])

	// Commands
	for _, op := range []string{"add", "append", "cas", "decr", "flushall", "flushre", "get", "gets", "incr", "metaget", "prepend", "replace", "touch", "set", "delete", "lease_get", "lease_set"} {
		key := "cmd_" + op
		ch <- prometheus.MustNewConstMetric(
			e.commands, prometheus.GaugeValue, e.parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandCount, prometheus.CounterValue, e.parse(s, key+"_count"), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandOut, prometheus.CounterValue, e.parse(s, key+"_out"), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandOutAll, prometheus.CounterValue, e.parse(s, key+"_out_all"), op)
	}

	ch <- prometheus.MustNewConstMetric(
		e.devNullRequests, prometheus.CounterValue, e.parse(s, "dev_null_requests"))
	ch <- prometheus.MustNewConstMetric(
		e.duration, prometheus.GaugeValue, e.parse(s, "duration_us"))
	ch <- prometheus.MustNewConstMetric(
		e.fibersAllocated, prometheus.GaugeValue, e.parse(s, "fibers_allocated"))
	ch <- prometheus.MustNewConstMetric(
		e.fibersPoolSize, prometheus.GaugeValue, e.parse(s, "fibers_pool_size"))
	ch <- prometheus.MustNewConstMetric(
		e.proxyReqsProcessing, prometheus.GaugeValue, e.parse(s, "proxy_reqs_processing"))
	ch <- prometheus.MustNewConstMetric(
		e.proxyReqsWaiting, prometheus.GaugeValue, e.parse(s, "proxy_reqs_waiting"))

	// Config
	ch <- prometheus.MustNewConstMetric(
		e.configFailures, prometheus.CounterValue, e.parse(s, "config_failures"))
	ch <- prometheus.MustNewConstMetric(
		e.configLastAttempt, prometheus.GaugeValue, e.parse(s, "config_last_attempt"))
	ch <- prometheus.MustNewConstMetric(
		e.configLastSuccess, prometheus.GaugeValue, e.parse(s, "config_last_success"))

	// Request
	for _, op := range []string{"error", "replied", "sent", "success"} {
		key := "request_" + op
		ch <- prometheus.MustNewConstMetric(
			e.requests, prometheus.GaugeValue, e.parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.requestCount, prometheus.CounterValue, e.parse(s, key+"_count"), op)
	}

	// Result Reply
	// See ProxyRequestLogger.cpp
	for _, op := range []string{"busy", "connect_error", "connect_timeout", "data_timeout", "error", "local_error", "tko"} {
		key := "result_" + op
		ch <- prometheus.MustNewConstMetric(
			e.results, prometheus.GaugeValue, e.parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultCount, prometheus.CounterValue, e.parse(s, key+"_count"), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultAll, prometheus.GaugeValue, e.parse(s, key+"_all"), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultAllCount, prometheus.CounterValue, e.parse(s, key+"_all_count"), op)
	}

	// Clients
	ch <- prometheus.MustNewConstMetric(
		e.clients, prometheus.CounterValue, e.parse(s, "num_clients"))
	ch <- prometheus.MustNewConstMetric(
		e.numClientConnections, prometheus.GaugeValue, e.parse(s, "num_client_connections"))

	// Servers
	for _, op := range []string{"closed", "down", "new", "up"} {
		key := "num_servers_" + op
		ch <- prometheus.MustNewConstMetric(
			e.servers, prometheus.GaugeValue, e.parse(s, key), op)
	}

	// Process stats
	ch <- prometheus.MustNewConstMetric(
		e.cpuSeconds, prometheus.CounterValue, e.parse(s, "ps_user_time_sec")+e.parse(s, "ps_system_time_sec"))
	ch <- prometheus.MustNewConstMetric(e.residentMemory, prometheus.CounterValue, e.parse(s, "ps_rss"))
	ch <- prometheus.MustNewConstMetric(e.virtualMemory, prometheus.CounterValue, e.parse(s, "ps_vsize"))

	ch <- prometheus.MustNewConstMetric(e.asynclogRequests, prometheus.CounterValue, e.parse(s, "asynclog_requests"))
	ch <- prometheus.MustNewConstMetric(e.asynclogRequestsRate, prometheus.GaugeValue, e.parse(s, "asynclog_requests_rate"))
	ch <- prometheus.MustNewConstMetric(e.asynclogSpoolSuccessRate, prometheus.GaugeValue, e.parse(s, "asynclog_spool_success_rate"))

	if e.server_stats {
		// Per-server stats
		s1, err := getServerStats(conn)

		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect server stats from mcrouter", "err", err)
			return
		}

		for server, metrics := range s1 {
			ch <- prometheus.MustNewConstMetric(
				e.serverDuration, prometheus.GaugeValue, e.parse(metrics, "avg_latency_us"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverProxyReqsProcessing, prometheus.GaugeValue, e.parse(metrics, "pending_reqs"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverProxyInflightReqs, prometheus.GaugeValue, e.parse(metrics, "inflight_reqs"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverProxyRetransRatio, prometheus.GaugeValue, e.parse(metrics, "avg_retrans_ratio"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedStored, prometheus.CounterValue, e.parse(metrics, "stored"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedNotStored, prometheus.CounterValue, e.parse(metrics, "notstored"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedFound, prometheus.CounterValue, e.parse(metrics, "found"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedNotFound, prometheus.CounterValue, e.parse(metrics, "notfound"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedDeleted, prometheus.CounterValue, e.parse(metrics, "deleted"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedTouched, prometheus.CounterValue, e.parse(metrics, "touched"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedExists, prometheus.CounterValue, e.parse(metrics, "exists"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedRemoteError, prometheus.CounterValue, e.parse(metrics, "remote_error"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedConnectTimeout, prometheus.CounterValue, e.parse(metrics, "connect_timeout"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedTimeout, prometheus.CounterValue, e.parse(metrics, "timeout"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedSoftTKO, prometheus.GaugeValue, e.parse(metrics, "soft_tko"), server)
			ch <- prometheus.MustNewConstMetric(
				e.serverMemcachedHardTKO, prometheus.GaugeValue, e.parse(metrics, "hard_tko"), server)
		}
	}

	if e.admin_requests {
		version, err := getAdminRequest(conn, "__mcrouter__.version")
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect version", "err", err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.adminRequestVersion, prometheus.GaugeValue, 1, string(version))

		configAgeStr, err := getAdminRequest(conn, "__mcrouter__.config_age")
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect config age", "err", err)
			return
		}

		configAge, err := strconv.ParseFloat(string(configAgeStr), 64)
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect config age", "err", err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.adminRequestConfigAge, prometheus.GaugeValue, configAge)

		configFile, err := getAdminRequest(conn, "__mcrouter__.config_file")
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect config file", "err", err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.adminRequestConfigFile, prometheus.GaugeValue, 1, string(configFile))

		hostId, err := getAdminRequest(conn, "__mcrouter__.hostid")
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect host id", "err", err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.adminRequestHostId, prometheus.GaugeValue, 1, string(hostId))

		config_md5_digest, err := getAdminRequest(conn, "__mcrouter__.config_md5_digest")
		if err != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			level.Error(e.logger).Log("msg", "Failed to collect admin config_md5_digest from mcrouter", "err", err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			e.adminRequestConfigMD5Digest, prometheus.GaugeValue, 1, string(config_md5_digest))
	}
}

// Parse a string into a 64 bit float suitable for  Prometheus
func (e *Exporter) parse(stats map[string]string, key string) float64 {
	val, ok := stats[key]
	if !ok {
		return 0.0
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		level.Error(e.logger).Log("msg", "Failed to parse", "key", key, "stat", stats[key], "err", err)
		v = math.NaN()
	}
	return v
}

// Get stats from mcrouter using a basic TCP connection
func getStats(conn net.Conn) (map[string]string, error) {
	m := make(map[string]string)
	fmt.Fprintf(conn, "stats all\r\n")
	reader := bufio.NewReader(conn)

	// Iterate over the lines and extract the metric name and value(s)
	// example lines:
	// 	 [STAT version 37.0.0
	//	 [STAT commandargs --option1 value --flag2
	//	 END
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "END\r\n" {
			break
		}

		// Split the line into 3 components, anything after the metric name should
		// be considered as the metric value.
		result := strings.SplitN(line, " ", 3)
		value := strings.TrimRight(result[2], "\r\n")
		m[result[1]] = value
	}

	return m, nil
}

// Get detailed per-server stats from mcrouter using a basic TCP connection
func getServerStats(conn net.Conn) (map[string]map[string]string, error) {
	m := make(map[string]map[string]string)
	fmt.Fprintf(conn, "stats servers\r\n")
	reader := bufio.NewReader(conn)

	// Iterate over the lines and extract the metric name and value(s)
	// example lines:
	// 	 STAT 10.64.16.110:11211:ascii:plain:notcompressed-1000 avg_latency_us:302.991
	//        pending_reqs:0 inflight_reqs:0 avg_retrans_ratio:0 max_retrans_ratio:0
	//        min_retrans_ratio:0 up:5; deleted:4875 touched:33069 found:112675373
	//        notfound:3493823 notstored:149776 stored:3250883 exists:2653 remote_error:32
	//	 END
	// In the same line there are two type of info:
	// * per-server stats about latency, requests, etc.. up to the ';' - (ProxyDestinationBase states)
	// * per-server breakdown of the memcached responses (STORED, DELETED, etc..) - (Carbon results)
	//   The memcached responses are listed in carbon_result.thrift, and they are returned/appended only
	//   when > 0. Special logic needs to be implemented then to set the ones not diplayed to zero,
	//   since Prometheus doesn't really like metrics appearing/disappearing.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "END\r\n" {
			break
		}

		line_sanitized := strings.TrimRight(line, "\r\n")

		// Split the line in 2 macro-components
		result := strings.SplitN(line_sanitized, ";", 2)

		server_metrics := strings.Split(strings.Trim(result[0], " "), " ")

		// WARNING: the memcached's result states (from ';' onward) might not be
		// available all the times, since they are added only if some data is present.
		// There are use cases, like mcrouter just started without any commands processed,
		// that do not expose any result. The exporter needs to add some defensive code
		// to avoid panic states.
		memcached_responses_metrics := []string{}
		if len(result) == 2 {
			memcached_responses_metrics = strings.Split(strings.Trim(result[1], " "), " ")
		}

		// The server id is always the second element of the row (after STAT),
		// and it will be used as label in all the metrics
		server_id := server_metrics[1]
		m[server_id] = make(map[string]string)

		// The following for loops assume that the two lines to parse
		// have the format: metric1:value1 metric2:value2 etc..
		// There are some special cases:
		// 1) the server_id string is the only exception since it is composed
		//    by more than one ':'.
		// 2) 'soft_tko' and 'hard_tko' are server flags that appear only when one
		//     shard is marked with that state, so they can be present or not
		//     depending on the runtime environment. To keep a stable metric,
		//     their values are initialized to 0 and turned to 1 only when the
		//     flag is found.
		SOFT_TKO_STATE := "soft_tko"
		HARD_TKO_STATE := "hard_tko"
		m[server_id][SOFT_TKO_STATE] = "0"
		m[server_id][HARD_TKO_STATE] = "0"
		for i := 2; i < len(server_metrics); i++ {
			if server_metrics[i] == SOFT_TKO_STATE || server_metrics[i] == HARD_TKO_STATE {
				m[server_id][server_metrics[i]] = "1"
			} else {
				metric_value := strings.SplitN(server_metrics[i], ":", 2)
				if len(metric_value) == 2 {
					m[server_id][metric_value[0]] = metric_value[1]
				}
			}
		}

		// See carbon_result.thrift in mcrouter's codebase
		// and also https://github.com/facebook/mcrouter/wiki/Error-Handling
		memcached_states := []string{"deleted", "touched", "found", "notfound", "notstored", "stored",
			"exists", "timeout", "connect_timeout", "remote_error",
		}

		// Set all the metrics to zero to create a baseline. mcrouter reports only
		// the states that have values > 0, but Prometheus doesn't like metrics
		// appearing and disappearing.
		for _, state := range memcached_states {
			m[server_id][state] = "0"
		}

		for i := 0; i < len(memcached_responses_metrics); i++ {
			metric_value := strings.SplitN(memcached_responses_metrics[i], ":", 2)
			if len(metric_value) == 2 {
				m[server_id][metric_value[0]] = metric_value[1]
			}
		}
	}

	return m, nil
}

func getAdminRequest(conn net.Conn, request string) ([]byte, error) {
	var value []byte
	fmt.Fprintf(conn, "get "+request+"\r\n")
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "END\r\n" {
			break
		}

		it := new(item)
		size, err := scanGetResponseLine(line, it)
		if err != nil {
			return nil, err
		}

		it.Value = make([]byte, size+2)
		_, err = io.ReadFull(reader, it.Value)
		if err != nil {
			it.Value = nil
			return nil, err
		}
		if !bytes.HasSuffix(it.Value, []byte("\r\n")) {
			it.Value = nil
			return nil, fmt.Errorf("memcache: corrupt get result read")
		}

		value = it.Value[:size]
	}

	return value, nil
}

func scanGetResponseLine(line string, it *item) (size int, err error) {
	pattern := "VALUE %s %d %d\r\n"
	dest := []interface{}{&it.Key, &it.Flags, &size}
	n, err := fmt.Sscanf(line, pattern, dest...)
	if err != nil || n != len(dest) {
		return -1, fmt.Errorf("memcache: unexpected line in get response: %q", line)
	}
	return size, nil
}

type item struct {
	// Key is the Item's key (250 bytes maximum).
	Key string

	// Value is the Item's value.
	Value []byte

	// Flags are server-opaque flags whose semantics are entirely
	// up to the app.
	Flags uint32

	// Expiration is the cache expiration time, in seconds: either a relative
	// time from now (up to 1 month), or an absolute Unix epoch time.
	// Zero means the Item has no expiration time.
	Expiration int32
}

func main() {
	var (
		address       = flag.String("mcrouter.address", "localhost:5000", "mcrouter server TCP address (tcp4/tcp6) or UNIX socket path")
		timeout       = flag.Duration("mcrouter.timeout", time.Second, "mcrouter connect timeout.")
		showVersion   = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", ":9442", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		serverMetrics = flag.Bool("mcrouter.server_metrics", false, "Collect per-server metrics.")
		adminRequests = flag.Bool("mcrouter.admin_requests", false, "Collect admin requests.")
		logLevel      = flag.String(promlogflag.LevelFlagName, "info", promlogflag.LevelFlagHelp)
		logFormat     = flag.String(promlogflag.FormatFlagName, "logfmt", promlogflag.FormatFlagHelp)
	)
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("mcrouter_exporter"))
		os.Exit(0)
	}

	logConfig := &promlog.Config{}
	logConfig.Level = &promlog.AllowedLevel{}
	if err := logConfig.Level.Set(*logLevel); err != nil {
		fmt.Fprintln(os.Stdout, "Invalid log level:", *logLevel)
		os.Exit(1)
	}
	logConfig.Format = &promlog.AllowedFormat{}
	if err := logConfig.Format.Set(*logFormat); err != nil {
		fmt.Fprintln(os.Stdout, "Invalid log format:", *logFormat)
		os.Exit(1)
	}

	logger := promlog.New(logConfig)

	// Exporter collects metrics from a mcrouter server.

	level.Info(logger).Log("msg", "Starting mcrouter_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	prometheus.MustRegister(NewExporter(*address, *timeout, *serverMetrics, *adminRequests, logger))
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`<html>
             <head><title>Mcrouter Exporter</title></head>
             <body>
             <h1>Mcrouter Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	level.Info(logger).Log("msg", "Starting HTTP server on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
