package main

import (
	"fmt"
	"net"
	"reflect"
	"testing"
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
	ret := []byte("STAT pid 1\r\nSTAT parent_pid 0\r\nSTAT fibers_allocated 1\r\nEND\r\n")
	conn.Write(ret)
	conn.Close()
}

// Test parsing a sample stats result
func TestStatsParsing(t *testing.T) {
	// Create a dummy server to return stats info
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

	client, _ := net.Dial("tcp", "localhost:9213")
	stats, _ := getStats(client)
	expected := make(map[string]string)
	expected["pid"] = "1"
	expected["parent_pid"] = "0"
	expected["fibers_allocated"] = "1"
	if !reflect.DeepEqual(stats, expected) {
		t.Errorf("Failed to parse stats into string map:\nGot:%s\nExpected:%s", stats, expected)
	}

}
