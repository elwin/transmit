package scion

import (
	"encoding/gob"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"io"
	"net"
	"time"
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

	err = receiveHandshake(stream)
	if err != nil {
		return nil, err
	}

	return &Connection{
		stream,
		conn.LocalAddr(),
		conn.RemoteAddr(),
	}, nil
}

func receiveHandshake(rw io.ReadWriter) error {
	log.Debug("Waiting for handshake")

	var message Message
	var decoder = gob.NewDecoder(rw)
	err := decoder.Decode(&message)
	if err != nil {
		return err
	}

	log.Debug("Received Handshake", "msg", message.Data)

	// Avoid race condition
	time.Sleep(100 * time.Millisecond)

	var reply = Message{"Yo, this is server speaking"}
	var encoder = gob.NewEncoder(rw)
	err = encoder.Encode(&reply)
	if err != nil {
		return nil
	}

	log.Debug("Sent reply")


	return nil
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