package scion

import (
	"github.com/lucas-clemente/quic-go"
	"net"
)

type Connection struct {
	quic.Stream
	local   net.Addr
	remote  net.Addr
}

func (connection Connection) LocalAddr() net.Addr {
	return connection.local
}

func (connection Connection) RemoteAddr() net.Addr {
	return connection.remote
}