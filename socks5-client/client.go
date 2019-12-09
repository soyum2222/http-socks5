package socks5

import (
	"bufio"
	"encoding/binary"
	"errors"
	"net"
)

var (
	ErrNoSocks5      = errors.New("this proto is not socks5")
	ErrNoSupportAuth = errors.New("server dont has client support auth method")
)

const (
	SocksVersion = 5
)
const (
	AUTHNO      = 0
	AUTHGSSAPI  = 1
	AUTHUSERPWD = 2
	AUTHNONE    = 0xFF
)

const (
	CMDConnect = 1
	CMDBind    = 2
	CMDUDP     = 3
)
const (
	ATYPIpv4 = 1
	ATYPHost = 3
	ATYPIpv6 = 4
)

func GetSocks5Conn(socks5_addr string, atyp byte, addr string, port uint16) (net.Conn, error) {
	conn, err := FirstShakeHands(socks5_addr)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	conn, err = SecondHandshake(conn, atyp, addr, port)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	return conn, err

}

func SecondHandshake(conn net.Conn, atyp byte, addr string, port uint16) (net.Conn, error) {

	msg := make([]byte, 3)
	msg[0] = SocksVersion
	msg[1] = CMDConnect
	msg[2] = 0x00
	address := make([]byte, 0)
	var err error
	switch atyp {
	case ATYPIpv4:
		ip := net.ParseIP(addr)
		address, err = ip.MarshalText()
		if err != nil {
			return nil, err
		}
	case ATYPHost:
		addr := []byte(addr)

		address_len := len(addr)
		address = append(address, byte(address_len))
		address = append(address, addr...)

	}

	p := make([]byte, 2)
	binary.BigEndian.PutUint16(p, port)
	msg = append(msg, atyp)
	msg = append(msg, address...)
	msg = append(msg, p...)

	_, err = conn.Write(msg)
	if err != nil {
		return nil, err
	}

	server_reply := make([]byte, 4)
	_, err = conn.Read(server_reply)
	if err != nil {
		return nil, err
	}
	if server_reply[0] != SocksVersion {
		return nil, ErrNoSocks5
	}
	if server_reply[1] != 0 {
		return nil, errors.New("no succeeded")
	}
	if server_reply[2] != 0 {
		return nil, errors.New("no socks")
	}
	switch server_reply[3] {
	case ATYPIpv4:
		addr := make([]byte, 4)
		_, err = conn.Read(addr)
		if err != nil {
			return nil, err
		}
	case ATYPHost:
		addr := make([]byte, 16)
		_, err = conn.Read(addr)
		if err != nil {
			return nil, err
		}
	}

	_, err = conn.Read(make([]byte, 2))

	return conn, err
}

func FirstShakeHands(socks5_addr string) (net.Conn, error) {

	conn, err := net.Dial("tcp", socks5_addr)
	if err != nil {
		return conn, err
	}

	_, err = conn.Write([]byte{SocksVersion, 1, 0})
	if err != nil {
		return conn, err
	}

	reader := bufio.NewReader(conn)
	v, err := reader.ReadByte()
	if err != nil {
		return conn, err
	}
	if v != SocksVersion {
		return conn, ErrNoSocks5
	}

	method, err := reader.ReadByte()
	if err != nil {
		return conn, err
	}

	if method != 0 {
		return conn, ErrNoSupportAuth
	}

	return conn, nil

}
