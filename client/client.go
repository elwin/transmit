package client

import (
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"net"
)

type Sconnection struct {
	quic.Stream
	local   net.Addr
	remote  net.Addr
}

func (connection Sconnection) LocalAddr() net.Addr {
	return connection.local
}

func (connection Sconnection) RemoteAddr() net.Addr {
	return connection.remote
}


func Connect(l, r string) Sconnection {

	local, _ := snet.AddrFromString(l)
	remote, _ := snet.AddrFromString(r)


	sciond := sciond.GetDefaultSCIONDPath(&local.IA)
	dispatcher := ""

	err := snet.Init(local.IA, sciond, dispatcher)
	if err != nil {
		log.Error("Failed to initialize SCION", "msg", err)
	}

	err = squic.Init("", "")
	if err != nil {
		log.Error("Failed to initialize SQUIC", "msg", err)
	}

	session, err := squic.DialSCION(nil, local, remote, nil)
	if err != nil {
		log.Error("Unable to dial", "err", err)
	}

	stream, err := session.OpenStream()
	if err != nil {
		log.Error("Unable to open stream", "err", err)
	}

	return Sconnection{
		stream,
		local,
		remote,
	}
}
