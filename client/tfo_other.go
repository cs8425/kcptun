// +build !darwin,!freebsd,!linux

package main

import (
	"net"
	"time"
)

func bindTFO(listener *net.TCPListener) { }

func handleTFO(p1 net.Conn, target string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{}
	if timeout != 0 {
		dialer.Timeout = timeout
	}
	return dialer.Dial("tcp", target)
}

