package scion

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"net"
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

	return &Connection{
		stream,
		local,
		remote,
	}, nil
}
