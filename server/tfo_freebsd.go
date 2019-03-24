package main

import (
	"syscall"
	"net"
	"time"
)

const TCP_FASTOPEN int = 1025 // 0x401
const opt_LEVEL = syscall.IPPROTO_TCP //syscall.SOL_SOCKET // syscall.SOCK_STREAM

// not support TCP_FASTOPEN_CONNECT
func getTFODialer(timeout time.Duration) *net.Dialer {
	dialer := &net.Dialer{}
	if timeout != 0 {
		dialer.Timeout = timeout
	}
	return dialer
}

