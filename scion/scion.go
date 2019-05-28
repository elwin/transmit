package scion

import (
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
)

type Listener struct {
	local  snet.Addr
	socket quic.Listener
}

func Listen(addr snet.Addr) Listener {

	sciond := sciond.GetDefaultSCIONDPath(&addr.IA)
	dispatcher := ""

	err := snet.Init(addr.IA, sciond, dispatcher)
	if err != nil {
		log.Error("Failed to initialize SCION", "msg", err)
	}

	err = squic.Init("", "")
	if err != nil {
		log.Error("Failed to initialize SQUIC", "msg", err)
	}

	socket, err := squic.ListenSCION(nil, &addr, nil)
	if err != nil {
		log.Error("Unable to listen", "err", err)
	}

	return Listener{
		local:  addr,
		socket: socket,
	}
}

func (s Listener) Accept() quic.Session {

	session, err := s.socket.Accept()

	if err != nil {
		log.Error("Unable to accept quic session", "err", err)
	}

	log.Debug("QUIC session accepted", "src", session.RemoteAddr())

	return session
}
