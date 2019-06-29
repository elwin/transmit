package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/elwin/transmit/socket"

	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
)

type multisocket struct {
	sockets   []socket.DataSocket
	maxLength int
	sc        chan *striping.Segment
	done      chan bool
}

func NewMultisocket(sockets []socket.DataSocket, maxLength int) *multisocket {
	return &multisocket{
		sockets,
		maxLength,
		make(chan *striping.Segment),
		make(chan bool),
	}
}

func (socket *multisocket) Write(reader io.Reader) {

	for _, s := range socket.sockets {
		go dispatcher(s, socket.sc, socket.done)
	}

	eodc := striping.NewEODCSegment(uint64(len(socket.sockets)))
	socket.sc <- eodc

	curPos := 0

	for {

		buf := make([]byte, socket.maxLength)

		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}

		socket.sc <- striping.NewSegment(buf, curPos)

		curPos += n

		// time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(3 * time.Second)

	for range socket.sockets {
		socket.done <- true
		<-socket.done
	}

	return

}

func dispatcher(socket socket.DataSocket, sc chan *striping.Segment, done chan bool) {
	defer func() {
		eod := striping.NewHeader(0, 0, striping.BlockFlagEndOfData)
		err := sendHeader(socket, eod)
		if err != nil {
			log.Error("Something bad happened", "err", err)
		}
		done <- true

	}()

	for {

		select {
		case <-done:
			fmt.Println("Done")
			return
		case segment := <-sc:
			// log.Debug("New Segment", "hdr", segment.Header)
			err := send(socket, segment)
			if err != nil {
				log.Error("Something bad happened", "err", err)
			}
		}

	}
}

func sendHeader(socket socket.DataSocket, header *striping.Header) error {
	err := binary.Write(socket, binary.BigEndian, header)
	log.Debug("Wrote header", "hdr", header)
	if err != nil {
		return fmt.Errorf("failed to write header: %s", err)
	}

	return nil
}

func send(socket socket.DataSocket, segment *striping.Segment) error {
	err := sendHeader(socket, segment.Header)
	if err != nil {
		return err
	}

	if !segment.IsEODCount() {
		reader := bytes.NewReader(segment.Data)
		_, err := io.Copy(socket, reader)
		if err != nil {
			return fmt.Errorf("failed to write data: %s", err)
		}
	}

	return nil
}
