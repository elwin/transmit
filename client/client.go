package client

import (
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
)

type connection struct {
	local   snet.Addr
	remote  snet.Addr
	session quic.Session
}

func Connect(local, remote snet.Addr) connection {

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

	session, err := squic.DialSCION(nil, &local, &remote, nil)
	if err != nil {
		log.Error("Unable to dial", "err", err)
	}

	return connection{
		local:   local,
		remote:  remote,
		session: session,
	}
}

func (conn connection) OpenStream() quic.Stream {
	stream, err := conn.session.OpenStreamSync()
	if err != nil {
		log.Error("Unable to open stream", "err", err)
	}
	return stream
}

func (conn connection) Close() {
	err := conn.session.Close(nil)
	if err != nil {
		log.Error("Unable to close session", "err", err)
	}
}
