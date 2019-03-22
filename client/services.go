package main

import (
	"io"
	"log"
	"net"
	"net/url"
	"bytes"
	"strings"
	"fmt"

	"github.com/cs8425/smux"
)

func replyAndClose(p1 net.Conn, rpy int) {
	p1.Write([]byte{0x05, byte(rpy), 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	p1.Close()
}

// thanks: http://www.golangnote.com/topic/141.html
func handleSocks(p1 net.Conn, sess *smux.Session, quiet bool, buffersize int) {
	var b [320]byte
	n, err := p1.Read(b[:])
	if err != nil {
		Vlogln(quiet, "socks client read", p1, err)
		return
	}
	if b[0] != 0x05 { //only Socket5
		return
	}

	//reply: NO AUTHENTICATION REQUIRED
	p1.Write([]byte{0x05, 0x00})

	n, err = p1.Read(b[:])
	if b[1] != 0x01 { // 0x01: CONNECT
		replyAndClose(p1, 0x07) // X'07' Command not supported
		return
	}

	var backend string
	switch b[3] {
	case 0x01: //IP V4
		backend = net.IPv4(b[4], b[5], b[6], b[7]).String()
		if n != 10 {
			replyAndClose(p1, 0x07) // X'07' Command not supported
			return
		}
	case 0x03: //DOMAINNAME
		backend = string(b[5 : n-2]) //b[4] domain name length
	case 0x04: //IP V6
		backend = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		if n != 22 {
			replyAndClose(p1, 0x07) // X'07' Command not supported
			return
		}
	default:
		replyAndClose(p1, 0x08) // X'08' Address type not supported
		return
	}

	p2, err := sess.OpenStream()
	if err != nil {
		return
	}
	defer p2.Close()
	// send to proxy
	p2.Write(b[0:n])

	var b2 [10]byte
	n2, err := p2.Read(b2[:10])
	if n2 < 10 {
		Vlogln(quiet, "Dial err replay:", backend, n2)
		replyAndClose(p1, 0x03)
		return
	}
	if err != nil || b2[1] != 0x00 {
		Vlogln(quiet, "socks err to:", backend, n2, b2[1], err)
		replyAndClose(p1, int(b2[1]))
		return
	}

	Vlogln(quiet, "socks to:", backend)
	reply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p1.Write(reply) // reply OK
	cp(p1, p2, buffersize)
}

// thanks: http://www.golangnote.com/topic/141.html
func handleHttp(client net.Conn, sess *smux.Session, quiet bool, buffersize int) {
	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		Vlogln(quiet, "http client read err", client, err)
		return
	}
	var method, host, address string
	idx := bytes.IndexByte(b[:], '\n')
	if idx == -1 {
		Vlogln(quiet, "http client parse err", idx, client.RemoteAddr())
		return
	}
	fmt.Sscanf(string(b[:idx]), "%s%s", &method, &host)

	if strings.Index(host, "://") == -1 {
		host = "//" + host
	}
	hostPortURL, err := url.Parse(host)
	if err != nil {
		Vlogln(quiet, "Parse hostPortURL err:", client, hostPortURL, err)
		return
	}
	if strings.Index(hostPortURL.Host, ":") == -1 { // no port, default 80
		address = hostPortURL.Host + ":80"
	} else {
		address = hostPortURL.Host
	}


	p2, err := sess.OpenStream()
	if err != nil {
		return
	}
	defer p2.Close()

	Vlogln(quiet, "Dial to:", method, address)
	var target = append([]byte{0, 0, 0, 0x05}, []byte(address)...)
	p2.Write(target)

	var b2 [10]byte
	n2, err := p2.Read(b2[:10])
	if n2 < 10 {
		Vlogln(quiet, "Dial err replay:", address, n2)
		return
	}
	if err != nil || b2[1] != 0x00 {
		Vlogln(quiet, "Dial err:", address, n2, b2[1], err)
		return
	}

	if method == "CONNECT" {
		client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		p2.Write(b[:n])
	}

	cp(client, p2, buffersize)
}

func handleClient(sess *smux.Session, p1 net.Conn, quiet bool, buffersize int, serv string) {
	if !quiet {
		log.Println("stream opened")
		defer log.Println("stream closed")
	}

	defer p1.Close()
	switch serv {
	case "socks5":
		handleSocks(p1, sess, quiet, buffersize)

	case "http":
		handleHttp(p1, sess, quiet, buffersize)

	default:
		p2, err := sess.OpenStream()
		if err != nil {
			return
		}
		defer p2.Close()
		cp(p1, p2, buffersize)
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

