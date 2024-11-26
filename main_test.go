package main

import (
	"fmt"
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	TEST_MCROUTER = "localhost:5000"
)

// Expect incoming message: stats all
// Return Example stats
func handleRequestStats(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	ret := []byte("STAT start_time 1\r\nSTAT version 0.0\r\nSTAT fibers_allocated 1\r\nSTAT fibers_pool_size 1\r\nSTAT commandargs --debug-fifo-root /var/lib/mcrouter/fifos --test-mode\r\nEND\r\n")
	conn.Write(ret)
	conn.Close()
}

// Expect incoming message: stats all
// Return Example stats
func handleRequestServerStats(conn net.Conn, full bool) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	ret := []byte("STAT 10.1.1.1:11211:ascii:plain:notcompressed-1000 avg_latency_us:302.991 pending_reqs:0 inflight_reqs:0 avg_retrans_ratio:0 max_retrans_ratio:0 min_retrans_ratio:0 up:5\r\n" +
		"STAT 10.1.1.2:11211:ascii:plain:notcompressed-1000 avg_latency_us:303.4 pending_reqs:0 inflight_reqs:0 avg_retrans_ratio:2 max_retrans_ratio:10 min_retrans_ratio:0 up:5\r\n" +
		"END\r\n")
	if full {
		ret = []byte("STAT 10.1.1.1:11211:ascii:plain:notcompressed-1000 avg_latency_us:302.991 pending_reqs:0 inflight_reqs:0 avg_retrans_ratio:0 max_retrans_ratio:0 min_retrans_ratio:0 up:5 soft_tko; " +
			"deleted:4875 touched:33069 found:112675373 notfound:3493823 notstored:149776 stored:3250883 exists:2653 remote_error:32\r\n" +
			"STAT 10.1.1.2:11211:ascii:plain:notcompressed-1000 avg_latency_us:303.4 pending_reqs:0 inflight_reqs:0 avg_retrans_ratio:2 max_retrans_ratio:10 min_retrans_ratio:0 up:5 hard_tko; " +
			"deleted:42 touched:3304 found:1175373 notfound:33823 notstored:0 stored:3250883 remote_error:55\r\n" +
			"END\r\n")
	}
	conn.Write(ret)
	conn.Close()
}

func handleAdminRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	var ret []byte

	command := string(buf[:n])

	switch command {
	case "get __mcrouter__.version\r\n":
		ret = []byte("VALUE __mcrouter__.version 0 15\r\n40.0.0 mcrouter\r\nEND\r\n")
	case "get __mcrouter__.config_age\r\n":
		ret = []byte("VALUE __mcrouter__.config_age 0 7\r\n1040469\r\nEND\r\n")
	case "get __mcrouter__.config_file\r\n":
		ret = []byte("VALUE __mcrouter__.config_file 0 32\r\n/opt/mcrouter/config/config.json\r\nEND\r\n")
	case "get __mcrouter__.hostid\r\n":
		ret = []byte("VALUE __mcrouter__.hostid 0 10\r\n4079863250\r\nEND\r\n")
	case "get __mcrouter__.config_md5_digest\r\n":
		ret = []byte("VALUE __mcrouter__.config_md5_digest 0 32\r\n2aa22ce671e9fdf6a7bb762f9a6cb0cc\r\nEND\r\n")
	default:
		ret = []byte("unknown command\r\nEND\r\n")
	}

	conn.Write(ret)
	conn.Close()
}

func TestStatsParsing(t *testing.T) {
	Convey("Given a remote mcrouter stats endpoint", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleRequestStats(server)
		}()
		Convey("When scraped by our client", func() {
			stats, err := getStats(client)
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the stats into a string map", func() {
				expected := make(map[string]string)
				expected["start_time"] = "1"
				expected["version"] = "0.0"
				expected["fibers_allocated"] = "1"
				expected["fibers_pool_size"] = "1"
				expected["commandargs"] = "--debug-fifo-root /var/lib/mcrouter/fifos --test-mode"
				So(stats, ShouldResemble, expected)
			})
		})

	})
}

func TestFullServerStatsParsing(t *testing.T) {
	Convey("Given a remote mcrouter stats server endpoint", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleRequestServerStats(server, true)
		}()

		Convey("When scraped by our client", func() {
			stats, err := getServerStats(client)
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the stats into a string map", func() {
				expected := make(map[string]map[string]string)
				expected["10.1.1.1:11211:ascii:plain:notcompressed-1000"] = map[string]string{
					"avg_latency_us": "302.991", "avg_retrans_ratio": "0", "connect_timeout": "0", "deleted": "4875",
					"exists": "2653", "found": "112675373", "inflight_reqs": "0", "max_retrans_ratio": "0", "min_retrans_ratio": "0",
					"notfound": "3493823", "notstored": "149776", "pending_reqs": "0", "remote_error": "32", "stored": "3250883",
					"timeout": "0", "soft_tko": "1", "hard_tko": "0", "touched": "33069", "up": "5",
				}
				expected["10.1.1.2:11211:ascii:plain:notcompressed-1000"] = map[string]string{
					"avg_latency_us": "303.4", "avg_retrans_ratio": "2", "connect_timeout": "0", "deleted": "42", "exists": "0",
					"found": "1175373", "inflight_reqs": "0", "max_retrans_ratio": "10", "min_retrans_ratio": "0", "notfound": "33823",
					"notstored": "0", "pending_reqs": "0", "remote_error": "55", "stored": "3250883", "timeout": "0", "soft_tko": "0",
					"hard_tko": "1", "touched": "3304", "up": "5",
				}
				So(stats, ShouldResemble, expected)
			})
		})

	})
}

func TestServerStatsParsingAfterMcrouterBootstrap(t *testing.T) {
	Convey("Given a remote mcrouter (without any commands processed yet) stats server server endpoint", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleRequestServerStats(server, false)
		}()

		Convey("When scraped by our client", func() {
			stats, err := getServerStats(client)
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the stats into a string map", func() {
				expected := make(map[string]map[string]string)
				expected["10.1.1.1:11211:ascii:plain:notcompressed-1000"] = map[string]string{
					"avg_latency_us": "302.991", "avg_retrans_ratio": "0", "connect_timeout": "0", "deleted": "0",
					"exists": "0", "found": "0", "inflight_reqs": "0", "max_retrans_ratio": "0",
					"min_retrans_ratio": "0", "notfound": "0", "notstored": "0", "pending_reqs": "0",
					"remote_error": "0", "stored": "0", "timeout": "0", "soft_tko": "0", "hard_tko": "0", "touched": "0",
					"up": "5",
				}
				expected["10.1.1.2:11211:ascii:plain:notcompressed-1000"] = map[string]string{
					"avg_latency_us": "303.4", "avg_retrans_ratio": "2", "connect_timeout": "0", "deleted": "0", "exists": "0",
					"found": "0", "inflight_reqs": "0", "max_retrans_ratio": "10", "min_retrans_ratio": "0", "notfound": "0",
					"notstored": "0", "pending_reqs": "0", "remote_error": "0", "stored": "0", "timeout": "0", "soft_tko": "0",
					"hard_tko": "0", "touched": "0", "up": "5",
				}
				So(stats, ShouldResemble, expected)
			})
		})

	})
}

func TestAdminRequestParsing(t *testing.T) {
	Convey("Given a remote mcrouter admin request server version", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleAdminRequest(server)
		}()

		Convey("When scraped by our client", func() {
			data, err := getAdminRequest(client, "__mcrouter__.version")
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the version", func() {
				expected := []byte("40.0.0 mcrouter")
				So(data, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a remote mcrouter admin request server config_age", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleAdminRequest(server)
		}()

		Convey("When scraped by our client", func() {
			data, err := getAdminRequest(client, "__mcrouter__.config_age")
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the config age", func() {
				expected := []byte("1040469")
				So(data, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a remote mcrouter admin request server config_file", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleAdminRequest(server)
		}()

		Convey("When scraped by our client", func() {
			data, err := getAdminRequest(client, "__mcrouter__.config_file")
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the config file", func() {
				expected := []byte("/opt/mcrouter/config/config.json")
				So(data, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a remote mcrouter admin request server hostid", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleAdminRequest(server)
		}()

		Convey("When scraped by our client", func() {
			data, err := getAdminRequest(client, "__mcrouter__.hostid")
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the hostid", func() {
				expected := []byte("4079863250")
				So(data, ShouldResemble, expected)
			})
		})
	})

	Convey("Given a remote mcrouter admin request server config_md5_digest", t, func() {
		server, client := net.Pipe()
		go func() {
			go handleAdminRequest(server)
		}()

		Convey("When scraped by our client", func() {
			data, err := getAdminRequest(client, "__mcrouter__.config_md5_digest")
			if err != nil {
				t.Fatal(err)
			}
			Convey("It should parse the config_md5_digest", func() {
				expected := []byte("2aa22ce671e9fdf6a7bb762f9a6cb0cc")
				So(data, ShouldResemble, expected)
			})
		})
	})
}
