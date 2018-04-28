package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	namespace = "mcrouter"
)

// Exporter collects metrics from a mcrouter server.
type Exporter struct {
	conn *net.Conn

	up                  *prometheus.Desc
	startTime           *prometheus.Desc
	version             *prometheus.Desc
	commandArgs         *prometheus.Desc
	commands            *prometheus.Desc
	commandCount        *prometheus.Desc
	commandOut          *prometheus.Desc
	commandOutFailover  *prometheus.Desc
	commandOutShadow    *prometheus.Desc
	commandOutAll       *prometheus.Desc
	commandOutCount     *prometheus.Desc
	configFailures      *prometheus.Desc
	configLastAttempt   *prometheus.Desc
	configLastSuccess   *prometheus.Desc
	devNullRequests    *prometheus.Desc
	duration            *prometheus.Desc
	fibersAllocated     *prometheus.Desc
	proxyReqsProcessing *prometheus.Desc
	proxyReqsWaiting    *prometheus.Desc
	requests            *prometheus.Desc
	requestCount        *prometheus.Desc
	results             *prometheus.Desc
	resultCount         *prometheus.Desc
	resultFailover      *prometheus.Desc
	resultShadow        *prometheus.Desc
	resultAll           *prometheus.Desc
	resultAllCount      *prometheus.Desc
	clients             *prometheus.Desc
	servers             *prometheus.Desc
	cpuSeconds          *prometheus.Desc
	residentMemory      *prometheus.Desc
	virtualMemory       *prometheus.Desc
	asynclogRequests    *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter(server string, timeout time.Duration) *Exporter {
	conn, _ := net.DialTimeout("tcp", server, timeout)
	return &Exporter{
		conn: &conn,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the mcrouter server be reached.",
			nil,
			nil,
		),
		startTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_time_seconds"),
			"The timestamp of mcrouter daemon start.",
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
			"Command args used.",
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
			"How long ago (in seconds) mcrouter has reconfigured.",
			nil,
			nil,
		),
		configLastAttempt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "config_last_attempt"),
			"How long ago (in seconds) mcrouter has reconfigured.",
			nil,
			nil,
		),
		configLastSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "config_last_success"),
			"How long ago (in seconds) mcrouter has reconfigured.",
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
			"Number of connected clients.",
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
	ch <- e.servers
	ch <- e.cpuSeconds
	ch <- e.residentMemory
	ch <- e.virtualMemory
	ch <- e.asynclogRequests
}

// Collect fetches the statistics from the configured mcrouter server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	s, err := getStats(*e.conn)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		log.Errorf("Failed to collect stats from mcrouter: %s", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	// Parse basic stats
	ch <- prometheus.MustNewConstMetric(e.startTime, prometheus.CounterValue, parse(s, "start_time"))
	ch <- prometheus.MustNewConstMetric(e.version, prometheus.GaugeValue, 1, s["version"])
	ch <- prometheus.MustNewConstMetric(e.commandArgs, prometheus.GaugeValue, 1, s["commandargs"])

	// Commands
	for _, op := range []string{"add", "append", "cas", "decr", "flushall", "flushre", "get", "incr", "metaget", "prepend", "replace", "touch", "set", "delete", "lease_get", "lease_set"} {
		key := "cmd_" + op
		ch <- prometheus.MustNewConstMetric(
			e.commands, prometheus.GaugeValue, parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandCount, prometheus.CounterValue, parse(s, key+"_count"), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandOut, prometheus.CounterValue, parse(s, key+"_out"), op)
		ch <- prometheus.MustNewConstMetric(
			e.commandOutAll, prometheus.CounterValue, parse(s, key+"_out_all"), op)
	}

	ch <- prometheus.MustNewConstMetric(
		e.devNullRequests, prometheus.CounterValue, parse(s, "dev_null_requests"))
	ch <- prometheus.MustNewConstMetric(
		e.duration, prometheus.GaugeValue, parse(s, "duration_us"))
	ch <- prometheus.MustNewConstMetric(
		e.fibersAllocated, prometheus.GaugeValue, parse(s, "fibers_allocated"))
	ch <- prometheus.MustNewConstMetric(
		e.proxyReqsProcessing, prometheus.GaugeValue, parse(s, "proxy_reqs_processing"))
	ch <- prometheus.MustNewConstMetric(
		e.proxyReqsWaiting, prometheus.GaugeValue, parse(s, "proxy_reqs_waiting"))

	// Config
	ch <- prometheus.MustNewConstMetric(
		e.configFailures, prometheus.CounterValue, parse(s, "config_failures"))
	ch <- prometheus.MustNewConstMetric(
		e.configLastAttempt, prometheus.GaugeValue, parse(s, "config_last_attempt"))
	ch <- prometheus.MustNewConstMetric(
		e.configLastSuccess, prometheus.GaugeValue, parse(s, "config_last_success"))

	// Request
	for _, op := range []string{"error", "replied", "sent", "success"} {
		key := "request_" + op
		ch <- prometheus.MustNewConstMetric(
			e.requests, prometheus.GaugeValue, parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.requestCount, prometheus.CounterValue, parse(s, key+"_count"), op)
	}

	// Result Reply
	// See ProxyRequestLogger.cpp
	for _, op := range []string{"busy", "connect_error", "connect_timeout", "data_timeout", "error", "local_error", "tko"} {
		key := "result_" + op
		ch <- prometheus.MustNewConstMetric(
			e.results, prometheus.GaugeValue, parse(s, key), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultCount, prometheus.CounterValue, parse(s, key+"_count"), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultAll, prometheus.GaugeValue, parse(s, key+"_all"), op)
		ch <- prometheus.MustNewConstMetric(
			e.resultAllCount, prometheus.CounterValue, parse(s, key+"_all_count"), op)
	}

	// Clients
	ch <- prometheus.MustNewConstMetric(
		e.clients, prometheus.CounterValue, parse(s, "num_clients"))

	// Servers
	for _, op := range []string{"closed", "down", "new", "up"} {
		key := "num_servers_" + op
		ch <- prometheus.MustNewConstMetric(
			e.servers, prometheus.GaugeValue, parse(s, key), op)
	}

	// Process stats
	ch <- prometheus.MustNewConstMetric(
		e.cpuSeconds, prometheus.CounterValue, parse(s, "ps_user_time_sec")+parse(s, "ps_system_time_sec"))
	ch <- prometheus.MustNewConstMetric(e.residentMemory, prometheus.CounterValue, parse(s, "ps_rss"))
	ch <- prometheus.MustNewConstMetric(e.virtualMemory, prometheus.CounterValue, parse(s, "ps_vsize"))

	ch <- prometheus.MustNewConstMetric(e.asynclogRequests, prometheus.CounterValue, parse(s, "asynclog_requests"))
}

// Parse a string into a 64 bit float suitable for  Prometheus
func parse(stats map[string]string, key string) float64 {
	v, err := strconv.ParseFloat(stats[key], 64)
	if err != nil {
		log.Errorf("Failed to parse %s %q: %s", key, stats[key], err)
		v = math.NaN()
	}
	return v
}

// Get stats from mcrouter using a basic TCP connection
func getStats(conn net.Conn) (map[string]string, error) {
	m := make(map[string]string)
	fmt.Fprintf(conn, "stats all\r\n")
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "END\r\n" {
			break
		}

		result := strings.Split(line, " ")
		value := strings.TrimRight(result[2], "\r\n")
		m[result[1]] = value
	}

	return m, nil
}

func main() {
	var (
		address       = flag.String("mcrouter.address", "localhost:5000", "mcrouter server address.")
		timeout       = flag.Duration("mcrouter.timeout", time.Second, "mcrouter connect timeout.")
		showVersion   = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", ":9151", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("mcrouter_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting mcrouter_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(NewExporter(*address, *timeout))
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Mcrouter Exporter</title></head>
             <body>
             <h1>Mcrouter Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Infoln("Starting HTTP server on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
