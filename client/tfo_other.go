// +build !darwin,!freebsd,!linux

package main

import (
	"net"
	"time"
)

func bindTFO(listener *net.TCPListener) { }

func getTFODialer(timeout time.Duration) *net.Dialer {
	dialer := &net.Dialer{}
	if timeout != 0 {
		dialer.Timeout = timeout
	}
	return dialer
}

