package scion

import (
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/snet"
)

type Listener interface {
	Addr() snet.Addr
	Close() error
	Accept() (Conn, error)
}

var _ Listener = Slistener{}

type Slistener struct {
	quicListener quic.Listener
	address      snet.Addr
}

func (listener Slistener) Addr() snet.Addr {
	return listener.address
}

func (listener Slistener) Close() error {
	return listener.quicListener.Close()
}

func (listener Slistener) Accept() (Conn, error) {
	conn, err := listener.quicListener.Accept()
	if err != nil {
		return nil, fmt.Errorf("couldn't accept SQUIC connection: %s", err)
	}

	stream, err := conn.AcceptStream()

	err = receiveHandshake(stream)
	if err != nil {
		return nil, err
	}

	return &Connection{
		stream,
		snet.Addr{},
		snet.Addr{},
	}, nil
}
