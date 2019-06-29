package ftp

import (
	"github.com/elwin/transmit/scion"
	"github.com/scionproto/scion/go/lib/log"
)

type rmsocket struct {
	sockets []scion.Conn
	done    chan bool
}

func NewReadMultisocket(sockets []scion.Conn) *rmsocket {
	return &rmsocket{sockets, make(chan bool)}
}

func (socket *rmsocket) _Read() []byte {
	transmission := NewTransmission()

	for {

		finished, err := transmission.ProcessBlock(socket.sockets[0], 0)
		if err != nil {
			log.Error("failed to process block", "err", err)
		}

		if finished {
			return transmission.getData()
		}

	}
}

func (socket *rmsocket) Read() []byte {

	transmission := NewTransmission()

	for i, s := range socket.sockets {
		go dispatcher(s, transmission, socket.done, i)
	}

	for range socket.sockets {
		<-socket.done
	}

	return transmission.getData()

}

func dispatcher(socket scion.Conn, transmission *transmission, done chan bool, i int) {
	defer func() {
		done <- true
	}()

	for {

		finished, err := transmission.ProcessBlock(socket, i)
		if err != nil {
			log.Error("failed to process block", "err", err)
		}

		if finished {
			return
		}

	}

}
