package main

import (
	"syscall"
)

const TCP_FASTOPEN int = 1025 // 0x401
const opt_LEVEL = syscall.IPPROTO_TCP //syscall.SOL_SOCKET // syscall.SOCK_STREAM
// not support TCP_FASTOPEN_CONNECT

