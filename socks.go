package nttp

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	socks5   = 0x05
	reserved = 0x00

	methodNoAuth           = 0x00
	methodUsernamePassword = 0x02 // This is actually not used...
	methodNotAcceptable    = 0xFF

	commandConnect   = 0x01
	commandBind      = 0x02 // TODO: Support this
	commandAssociate = 0x03 // 		 and this as well!

	replySuccess             = 0x00
	replyGeneralFailure      = 0x01
	replyAddressNotSupported = 0x08
	replyCommandNotSupported = 0x07

	addressIPv4       = 0x01
	addressDomainName = 0x03
	addressIPv6       = 0x04
)

func pipeBetween(a io.ReadWriter, b io.ReadWriter) {
	done := make(chan bool, 2)
	go func() { _, _ = io.Copy(a, b); done <- true }()
	go func() { _, _ = io.Copy(b, a); done <- true }()
	<-done
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Println("Error when closing reader:", err)
	}
}

func ListenAsClient(listen, remote string) {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer Close(l)
	log.Println("Listening for SOCKS5 connection on", l.Addr())
	for {
		connLocal, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			defer Close(connLocal)
			log.Println("Accepted SOCKS client connection from", connLocal.RemoteAddr().String())
			connServer, err := net.Dial("tcp", remote)
			if err != nil {
				log.Println(err)
				return
			}
			defer Close(connServer)
			log.Println("Established NTT relay from", connLocal.RemoteAddr().String(), "to", connServer.RemoteAddr().String())
			pipeBetween(connLocal, newNTTReadWriter(connServer))
			log.Println("Finished relay with", connLocal.RemoteAddr().String())
		}()
	}
}

func ListenAsServer(listen, self string) {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer Close(l)
	log.Println("Listening for NTTP client connection on", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			log.Println("Accepted socks relay connection from", conn.RemoteAddr().String())
			defer Close(conn)
			handleSocks5Conn(conn, self)
		}()
	}
}

func hasMethod(msg []byte, method byte) bool {
	n := int(msg[1])
	for _, m := range msg[2 : 2+n] {
		if m == method {
			return true
		}
	}
	return false
}

func encodeAddr(addr string) []byte {
	ip := net.ParseIP(addr)
	if ip != nil {
		switch len(ip) {
		case net.IPv4len:
			return append([]byte{addressIPv4}, ip...)
		case net.IPv6len:
			return append([]byte{addressIPv6}, ip...)
		}
	}
	return append([]byte{addressDomainName, byte(len(addr))}, addr...)
}

func encodeAddrAndPort(addr string, port int) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(port))
	return append(encodeAddr(addr), buf...)
}

func handleSocks5Conn(conn net.Conn, selfAddr string) {
	buf := make([]byte, 2*BlockDataSize)
	s := newNTTReadWriter(conn)
	// Checks the greeting message from the client
	n, err := s.Read(buf)
	if err == nil {
		if n < 3 {
			err = fmt.Errorf("greeting message too short")
		} else if buf[0] != socks5 {
			err = fmt.Errorf("not SOCKS5 protocol")
		}
	}
	if err != nil {
		log.Println("Failed to greet with client", conn.RemoteAddr(), ":", err)
		return
	}

	// We use no auth by default
	if !hasMethod(buf, methodNoAuth) {
		log.Println("No proper methods for request from", conn.RemoteAddr())
		_, _ = s.Write([]byte{socks5, methodNotAcceptable})
		return
	}

	// And we send our method selection message
	_, err = s.Write([]byte{socks5, methodNoAuth})
	if err != nil {
		log.Println("Failed to select method from", conn.RemoteAddr(), ":", err)
		return
	}

	// Do something with request
	n, err = s.Read(buf)
	if err == nil {
		if n < 10 {
			err = fmt.Errorf("request message too short")
		} else if buf[0] != socks5 {
			err = fmt.Errorf("not SOCKS5 protocol")
		}
	}
	if err != nil {
		log.Println("Failed to handle request from", conn.RemoteAddr(), ":", err)
		return
	}

	// Helper function for sending reply
	reply := func(rep byte) {
		_, _ = s.Write([]byte{socks5, rep, reserved, addressIPv4,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	// Resolve address
	var dstAddr string
	switch buf[3] {
	case addressIPv4:
		dstAddr = net.IP(buf[4 : 4+net.IPv4len]).String()
	case addressIPv6:
		dstAddr = net.IP(buf[4 : 4+net.IPv6len]).String()
	case addressDomainName:
		dstAddr = string(buf[5 : n-2])
	default:
		reply(replyAddressNotSupported)
		return
	}
	dstPort := int(binary.BigEndian.Uint16(buf[n-2:]))

	switch buf[1] {
	case commandConnect:
		dst, err := net.Dial("tcp", net.JoinHostPort(dstAddr, strconv.Itoa(dstPort)))
		if err != nil {
			log.Println("General failure when dialing", net.JoinHostPort(dstAddr, strconv.Itoa(dstPort)), ":", err)
			reply(replyGeneralFailure)
			return
		}
		defer Close(dst)
		reply(replySuccess)
		log.Println("Established NTT relay from", dst.RemoteAddr().String(), "to", conn.RemoteAddr().String())
		pipeBetween(s, dst)
	case commandBind:
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			reply(replyGeneralFailure)
			log.Println(err)
			return
		}
		defer Close(l)
		log.Println("Listening as told by BIND request on", l.Addr().String())
		listenPort := l.Addr().(*net.TCPAddr).Port
		_, _ = s.Write(append([]byte{socks5, replySuccess, reserved},
			encodeAddrAndPort(selfAddr, listenPort)...))
		coming, err := l.Accept()
		if err != nil {
			reply(replyGeneralFailure)
			log.Println(err)
			return
		}
		// TODO: Check here if comingAddr and comingPort matches with dstAddr and dstPort
		comingAddr := coming.RemoteAddr().(*net.TCPAddr).IP
		comingPort := coming.RemoteAddr().(*net.TCPAddr).Port
		_, _ = s.Write(append([]byte{socks5, replySuccess, reserved},
			encodeAddrAndPort(comingAddr.String(), comingPort)...))
		log.Println("Established NTT relay from", coming.RemoteAddr().String(), "to", conn.RemoteAddr().String())
		pipeBetween(s, coming)
	default:
		log.Println("Unsupported command from", conn.RemoteAddr(), ":", err)
		reply(replyCommandNotSupported)
	}
}
