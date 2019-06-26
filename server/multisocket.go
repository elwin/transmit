package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"io"
)

var _ io.Writer = &multisocket{}

type multisocket struct {
	sockets   []DataSocket
	maxLength int
}

func NewMultisocket(sockets []DataSocket, maxLength int) io.Writer {
	return &multisocket{sockets, maxLength}
}

func (socket *multisocket) Write(p []byte) (n int, err error) {

	sc := make(chan *striping.Segment)
	done := make(chan bool)

	for _, s := range socket.sockets {
		go dispatcher(s, sc, done)
	}

	eodc := striping.NewEODCSegment(uint64(len(socket.sockets)))
	sc <- eodc

	curPos := 0

	for {

		endPos := curPos + socket.maxLength
		if endPos > len(p) {
			endPos = len(p)
		}

		curData := p[curPos:endPos]

		sc <- striping.NewSegment(curData, curPos)

		curPos = endPos

		if curPos >= len(p) {
			break
		}
	}

	for range socket.sockets {
		done <- true
		<-done
	}

	return curPos, nil
}

func dispatcher(socket DataSocket, sc chan *striping.Segment, done chan bool) {
	defer func() { done <- true }()
	for {

		select {
		case <-done:
			log.Debug("Done")
			eod := striping.NewHeader(0, 0, striping.BlockFlagEndOfData)
			err := sendHeader(socket, eod)
			if err != nil {
				log.Error("Something bad happened", "err", err)
			}
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

func sendHeader(socket DataSocket, header *striping.Header) error {
	err := binary.Write(socket, binary.BigEndian, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %s", err)
	}

	return nil
}

func send(socket DataSocket, segment *striping.Segment) error {
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
