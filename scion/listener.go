package scion

import (
	"encoding/binary"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"strings"
)

type Listener interface {
	Addr() snet.Addr
	Close() error
	Accept() (Conn, error)
}

var _ Listener = ScionListener{}

type ScionListener struct {
	quicListener quic.Listener
	local        snet.Addr
}

func (listener ScionListener) Addr() snet.Addr {
	return listener.local
}

func (listener ScionListener) Close() error {
	return listener.quicListener.Close()
}

func (listener ScionListener) Accept() (Conn, error) {
	session, err := listener.quicListener.Accept()
	if err != nil {
		return nil, fmt.Errorf("couldn't accept SQUIC connection: %s", err)
	}

	remote := session.RemoteAddr().String()
	remote = strings.Split(remote, " ")[0]

	remoteAddr, err := snet.AddrFromString(remote)
	if err != nil {
		return nil, err
	}

	stream, err := session.AcceptStream()

	err = receiveHandshake(stream)
	if err != nil {
		return nil, err
	}

	return NewConnection(stream, listener.local, *remoteAddr), nil
}

func receiveHandshake(rw io.ReadWriter) error {

	msg := make([]byte, 1)
	err := binary.Read(rw, binary.BigEndian, msg)
	if err != nil {
		return err
	}

	return nil
}
