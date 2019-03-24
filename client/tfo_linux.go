// +build linux

package main

import (
	"syscall"
	"net"
	"time"
	"fmt"
)

const TCP_FASTOPEN int = 23 // 0x17
const TCP_FASTOPEN_CONNECT int = 30 // 0x1e

const opt_LEVEL = syscall.SOL_TCP

func getTFODialer(timeout time.Duration) *net.Dialer {
	dialer := &net.Dialer{}
	if timeout != 0 {
		dialer.Timeout = timeout
	}
	dialer.Control = func(network, address string, c syscall.RawConn) error {
		c.Control(func(fd uintptr) {
			err := syscall.SetsockoptInt(int(fd), opt_LEVEL, TCP_FASTOPEN_CONNECT, 1)
			if err != nil {
				fmt.Printf("Failed to set necessary TCP_FASTOPEN_CONNECT socket option: %s", err)
				return
			}
		})
		return nil
	}

	return dialer
}

