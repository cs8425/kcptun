// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"net"
	"syscall"
	"time"
	"fmt"
)

const TCP_FASTOPEN int = 23
const TCP_FASTOPEN_CONNECT int = 30

func bindTFO(listener *net.TCPListener) {
	rawconn, err := listener.SyscallConn()
	if err != nil {
		return
	}

	rawconn.Control(func(fd uintptr) {
		err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN, 1)
		if err != nil {
			fmt.Printf("Failed to set necessary TCP_FASTOPEN socket option: %s", err)
			return
		}
	})
}

func getTFODialer(timeout time.Duration) *net.Dialer {
	dialer := &net.Dialer{}
	if timeout != 0 {
		dialer.Timeout = timeout
	}
	dialer.Control = func(network, address string, c syscall.RawConn) error {
		c.Control(func(fd uintptr) {
			err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN_CONNECT, 1)
			if err != nil {
				fmt.Printf("Failed to set necessary TCP_FASTOPEN_CONNECT socket option: %s", err)
				return
			}
		})
		return nil
	}

	return dialer
}

