package scion

import (
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"net"
)

type Slistener struct {
	quic.Listener
}

type Sconnection struct {
	quic.Stream
	local net.Addr
	remote net.Addr
}

func (connection Sconnection) LocalAddr() net.Addr {
	return connection.local
}

func (connection Sconnection) RemoteAddr() net.Addr {
	return connection.remote
}

func (slistener Slistener) Addr() net.Addr {
	return slistener.Addr()
}

func (slistener Slistener) Close() error {
	return nil
}

func (slistener Slistener) Accept() (net.Conn, error) {
	conn, err := slistener.Listener.Accept()
	if err != nil {
		log.Error("Cannot accept connection", "msg", err)
	}

	stream, err := conn.AcceptStream()

	return Sconnection{stream, conn.LocalAddr(), conn.RemoteAddr()}, nil
}

func Listen(address string) Slistener {

	addr, _ := snet.AddrFromString(address)

	sciond := sciond.GetDefaultSCIONDPath(&addr.IA)
	dispatcher := ""

	err := snet.Init(addr.IA, sciond, dispatcher)
	if err != nil {
		log.Debug("Failed to initialize SCION", "msg", err)
	}

	err = squic.Init("", "")
	if err != nil {
		log.Error("Failed to initialize SQUIC", "msg", err)
	}

	listener, err := squic.ListenSCION(nil, addr, nil)
	if err != nil {
		log.Error("Unable to listen", "err", err)
	}


	return Slistener{listener}
}

/*

func (listener quic.Listener) Accept() quic.Session {

	session, err := listener.socket.Accept()

	if err != nil {
		log.Error("Unable to accept quic session", "err", err)
	}

	log.Debug("QUIC session accepted", "src", session.RemoteAddr())

	return session
}

*/
