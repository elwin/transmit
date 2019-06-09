package scion

import (
	"encoding/gob"
	"fmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"io"
	"net"
	"time"
)


func Dial(localAddr, remoteAddr string) (net.Conn, error) {

	local, err := snet.AddrFromString(localAddr)
	if err != nil {
		return nil, err
	}

	remote, err := snet.AddrFromString(remoteAddr)
	if err != nil {
		return nil, err
	}


	sciond := sciond.GetDefaultSCIONDPath(&local.IA)
	dispatcher := ""

	if snet.DefNetwork == nil {
		err = snet.Init(local.IA, sciond, dispatcher)
		if err != nil {
			return nil, fmt.Errorf("failted to initialize SCION: %s", err)
		}
	}

	err = squic.Init("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initilaze SQUIC: %s", err)
	}

	session, err := squic.DialSCION(nil, local, remote, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to dial %s: %s", remoteAddr, err)
	}

	stream, err := session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("unable to open stream: %s", err)
	}

	err = sendHandshake(stream)
	if err != nil {
		return nil, err
	}

	/*
	var decoder = gob.NewDecoder(stream)

	err = decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}

	log.Debug("Received response", "msg", msg.Data)


	 */

	return &Connection{
		stream,
		local,
		remote,
	}, nil
}

func sendHandshake(rw io.ReadWriter) error {
	var message = Message{"Hello World!"}
	var encoder = gob.NewEncoder(rw)
	err := encoder.Encode(&message)
	if err != nil {
		return err
	}

	log.Debug("Sent handshake")
	log.Debug("Waiting for reply")

	var reply Message
	var decoder = gob.NewDecoder(rw)
	err = decoder.Decode(&reply)
	if err != nil {
		return err
	}

	log.Debug("Received reply", "msg", reply.Data)

	// Avoid race condition
	time.Sleep(100 * time.Millisecond)

	return nil
}