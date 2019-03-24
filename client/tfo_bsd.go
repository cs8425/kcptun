// +build freebsd darwin

package main

import (
	"net"
	"os"
	"syscall"
	"time"
)

func handleTFO(p1 net.Conn, target string, timeout time.Duration) (net.Conn, error) {
	buf := make([]byte, 4096) // TODO: export setting
	p1.SetReadDeadline(time.Now().Add(50 * time.Millisecond)) // TODO: export setting
	n, err := p1.Read(buf) // try read first packet data
	if err, ok := err.(net.Error); ok && err.Timeout() { // read data timeout
		p1.SetReadDeadline(time.Time{})
		return net.DialTimeout("tcp", target, timeout)
		//return DialTFO(target, buf[:n])
	}
	if err != nil {
		return nil, err
	}
	p1.SetReadDeadline(time.Time{})

	return DialTFO(target, buf[:n])
}

func DialTFO(address string, data []byte) (net.Conn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	var sa syscall.Sockaddr
	family := syscall.AF_INET
	ip6 := raddr.IP.To16()
	ip4 := raddr.IP.To4()

	switch {
	case ip6 != nil: //IPv6
		family = syscall.AF_INET6
		sa6 := &syscall.SockaddrInet6{Port: raddr.Port, ZoneId: uint32(IP6ZoneToInt(raddr.Zone))}
		copy(sa6.Addr[:], ip6)
		sa = sa6

	case ip4 != nil:
		family = syscall.AF_INET
		sa4 := &syscall.SockaddrInet4{Port: raddr.Port}
		copy(sa4.Addr[:], ip4)
		sa = sa4
	}

	fd, err := syscall.Socket(family, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(fd)

	for {
		//err = syscall.Sendto(fd, data, syscall.MSG_FASTOPEN, sa) // linux version
		err = syscall.Sendto(fd, data, TCP_FASTOPEN, sa)
		if err == syscall.EAGAIN {
			continue
		}
		break
	}

	if _, ok := err.(syscall.Errno); ok {
		return nil, os.NewSyscallError("sendto", err)
	}

	return net.FileConn(os.NewFile(uintptr(fd), "TFO: " + raddr.String()))
}

// from: https://github.com/libp2p/go-sockaddr/blob/master/net/net.go
/*
Copyright (c) 2012 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
// IP6ZoneToInt converts an IP6 Zone net string to a unix int
// returns 0 if zone is ""
func IP6ZoneToInt(zone string) int {
	if zone == "" {
		return 0
	}
	if ifi, err := net.InterfaceByName(zone); err == nil {
		return ifi.Index
	}
	n, _, _ := dtoi(zone, 0)
	return n
}

// Bigger than we need, not too big to worry about overflow
const big = 0xFFFFFF

// Decimal to integer starting at &s[i0].
// Returns number, new offset, success.
func dtoi(s string, i0 int) (n int, i int, ok bool) {
	n = 0
	for i = i0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return 0, i, false
		}
	}
	if i == i0 {
		return 0, i, false
	}
	return n, i, true
}

