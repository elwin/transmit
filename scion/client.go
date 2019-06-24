package scion

import (
	"encoding/binary"
	"fmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"io"
)

func Dial(local, remote snet.Addr) (Conn, error) {
	sciond := sciond.GetDefaultSCIONDPath(&local.IA)
	dispatcher := ""

	if snet.DefNetwork == nil {
		err := snet.Init(local.IA, sciond, dispatcher)
		if err != nil {
			return nil, fmt.Errorf("failted to initialize SCION: %s", err)
		}
	}

	err := squic.Init("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initilaze SQUIC: %s", err)
	}

	session, err := squic.DialSCION(nil, &local, &remote, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to dial %s: %s", AddrToString(remote), err)
	}

	stream, err := session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("unable to open stream: %s", err)
	}

	err = sendHandshake(stream)
	if err != nil {
		return nil, err
	}

	return NewConnection(stream, local, remote), nil
}

func DialAddr(localAddr, remoteAddr string) (Conn, error) {

	local, err := snet.AddrFromString(localAddr)
	if err != nil {
		return nil, err
	}

	remote, err := snet.AddrFromString(remoteAddr)
	if err != nil {
		return nil, err
	}

	return Dial(*local, *remote)
}

func sendHandshake(rw io.ReadWriter) error {

	msg := []byte{200}

	binary.Write(rw, binary.BigEndian, msg)

	log.Debug("Sent handshake", "msg", msg)

	return nil
}
