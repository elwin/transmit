package scion

import (
	"encoding/gob"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"net"
)

type Listener struct {
	quic.Listener
}

func (listener Listener) Addr() net.Addr {
	return listener.Listener.Addr()
}

func (listener Listener) Close() error {
	return listener.Listener.Close()
}

func (listener Listener) Accept() (net.Conn, error) {
	conn, err := listener.Listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("couldn't accept SQUIC connection: %s", err)
	}

	stream, err := conn.AcceptStream()

	var msg Message

	var decoder = gob.NewDecoder(stream)

	err = decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}

	log.Debug("Received Handshake", "msg", msg.Data)

	/*
	var encoder = gob.NewEncoder(stream)

	msg.Data = "Thanks!"

	err = encoder.Encode(&msg)
	if err != nil {
		return nil, err
	}
	 */



	return &Connection{
		stream,
		conn.LocalAddr(),
		conn.RemoteAddr(),
	}, nil
}

func Listen(address string) (net.Listener, error) {

	addr, _ := snet.AddrFromString(address)

	// Initialize Network if uninitialized
	if snet.DefNetwork == nil {

		sciond := sciond.GetDefaultSCIONDPath(&addr.IA)
		dispatcher := ""

		err := snet.Init(addr.IA, sciond, dispatcher)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SCION: %s", err)
		}
	}

	// Initialize SQUIC
	err := squic.Init("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initilaze SQUIC: %s", err)
	}

	listener, err := squic.ListenSCION(nil, addr, nil)
	if err != nil {
		log.Error("Unable to listen", "err", err)
	}

	return &Listener{listener}, nil
}