package scion

import (
	"encoding/gob"
	"fmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"io"
	"time"
)

func receiveHandshake(rw io.ReadWriter) error {
	// log.Debug("Waiting for handshake")

	var message Message
	var decoder = gob.NewDecoder(rw)
	err := decoder.Decode(&message)
	if err != nil {
		return err
	}

	// log.Debug("Received Handshake", "msg", message.Data)

	// Avoid race condition
	time.Sleep(100 * time.Millisecond)

	var reply = Message{"Welcome to this server"}
	var encoder = gob.NewEncoder(rw)
	err = encoder.Encode(&reply)
	if err != nil {
		return nil
	}

	// log.Debug("Sent reply")

	return nil
}

func Listen(address string) (Listener, error) {

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

	return &ScionListener{
		listener,
		*addr,
	}, nil
}
