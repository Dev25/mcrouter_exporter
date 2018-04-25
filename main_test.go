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
func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	ret := []byte("STAT uptime 1\r\nSTAT version 0.0\r\nSTAT fibers_allocated 1\r\nEND\r\n")
	conn.Write(ret)
	conn.Close()
}

func TestStatsParsing(t *testing.T) {
	Convey("Given a remote mcrouter stats endpoint", t, func() {
		go func() {
			l, err := net.Listen("tcp", "localhost:9213")
			if err != nil {
				t.Fatal("Failed to create TCP Server:", err.Error())
			}
			defer l.Close()
			for {
				conn, err := l.Accept()
				if err != nil {
					t.Fatal("Error accepting:", err.Error())
				}
				go handleRequest(conn)
			}
		}()
		Convey("When scraped by our client", func() {
			client, _ := net.Dial("tcp", "localhost:9213")
			stats, _ := getStats(client)
			Convey("It should parse the stats into a string map", func() {
				expected := make(map[string]string)
				expected["uptime"] = "1"
				expected["version"] = "0.0"
				expected["fibers_allocated"] = "1"
				So(stats, ShouldResemble, expected)
			})
		})
	})
}
