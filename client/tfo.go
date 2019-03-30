// +build linux freebsd darwin

package main

import (
	"net"
	"syscall"
	"fmt"
)

func bindTFO(listener *net.TCPListener) {
	rawconn, err := listener.SyscallConn()
	if err != nil {
		return
	}

	rawconn.Control(func(fd uintptr) {
		err := syscall.SetsockoptInt(int(fd), opt_LEVEL, TCP_FASTOPEN, 1)
		if err != nil {
			fmt.Printf("Failed to set necessary TCP_FASTOPEN socket option: %s\n", err)
			return
		}
	})
}

