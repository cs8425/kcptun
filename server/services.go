package main

import (
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

func replyAndClose(p1 net.Conn, rpy int) {
	p1.Write([]byte{0x05, byte(rpy), 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	p1.Close()
}

func handleFast(p1 net.Conn, quiet bool, buffersize int, tfo bool) {
	var b [320]byte
	n, err := p1.Read(b[:])
	if err != nil {
		Vlogln(quiet, "[fast client read]", p1, err)
		return
	}
	// b[0:2] // ignore

	var host, port, backend string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //DOMAINNAME
		host = string(b[5 : n-2]) //b[4] domain name length
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	case 0x05: //DOMAINNAME + PORT
		backend = string(b[4 : n])
		goto CONN
	default:
		replyAndClose(p1, 0x08) // X'08' Address type not supported
		return
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	backend = net.JoinHostPort(host, port)


CONN:
	var p2 net.Conn
	if tfo {
		p2, err = getTFODialer(5 * time.Second).Dial("tcp", backend)
	} else {
		p2, err = net.DialTimeout("tcp", backend, 5 * time.Second)
	}
	if err != nil {
		Vlogln(quiet, "[err]", backend, err)

		switch t := err.(type) {
		case *net.AddrError:
			replyAndClose(p1, 0x03) // X'03' Network unreachable

		case *net.OpError:
			if t.Timeout() {
				replyAndClose(p1, 0x06) // X'06' TTL expired
			} else if t.Op == "dial" {
				replyAndClose(p1, 0x05) // X'05' Connection refused
			}

		default:
			//replyAndClose(p1, 0x03) // X'03' Network unreachable
			//replyAndClose(p1, 0x04) // X'04' Host unreachable
			replyAndClose(p1, 0x05) // X'05' Connection refused
			//replyAndClose(p1, 0x06) // X'06' TTL expired
		}
		return
	}
	defer p2.Close()

	Vlogln(quiet, "[got]", backend)
	reply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p1.Write(reply) // reply OK

	cp(p1, p2, buffersize)
	Vlogln(quiet, "[cls]", backend)
}

func handleClient(p1 net.Conn, quiet bool, buffersize int, serv string, target string, tfo bool) {
	if !quiet {
		log.Println("stream opened")
		defer log.Println("stream closed")
	}

	defer p1.Close()
	switch serv {
	default:
		fallthrough
	case "raw":

		var p2 net.Conn
		var err error
		if tfo {
			p2, err = getTFODialer(5 * time.Second).Dial("tcp", target)
		} else {
			p2, err = net.DialTimeout("tcp", target, 5 * time.Second)
		}
		if err != nil {
			Vlogln(quiet, "[connect err]", target, err)
			return
		}
		defer p2.Close()
		cp(p1, p2, buffersize)

	case "fast":
		handleFast(p1, quiet, buffersize, tfo)
	}

}

func cp(p1 net.Conn, p2 net.Conn, buffersize int) {
	defer p2.Close()

	// start tunnel
	p1die := make(chan struct{})
	buf1 := make([]byte, buffersize)
	go func() { io.CopyBuffer(p1, p2, buf1); close(p1die) }()

	p2die := make(chan struct{})
	buf2 := make([]byte, buffersize)
	go func() { io.CopyBuffer(p2, p1, buf2); close(p2die) }()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}
}

func Vlogln(quiet bool, v ...interface{}) {
	if !quiet {
		log.Println(v...)
	}
}

