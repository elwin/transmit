package socket

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"io"
)

// Currently only able to send data
// WriteTo doesn't work correctly

type multisocket struct {
	sockets   []DataSocket
	maxLength int
}

func (socket *multisocket) WriteTo(w io.Writer) (n int64, err error) {

	segmentChannel := make(chan *striping.Segment)
	done := make(chan struct{})

	for _, s := range socket.sockets {
		go reader(s, segmentChannel, done)
	}

	written := 0
	eodc := -1
	finished := 0
	queue := striping.NewSegmentQueue()

	for {

		// Listen to children
		select {
		case segment := <-segmentChannel:
			if segment.IsEODCount() {
				eodc = segment.GetEODCount()
			} else {
				queue.Push(segment)
			}

		case <-done:
			finished += 1
			if finished == eodc {
				return int64(written), nil
			}
		}

		// Write from Queue if possible

	}

}

func reader(socket DataSocket, sc chan *striping.Segment, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	for {

		segment, err := receiveNextSegment(socket)
		if err != nil {
			log.Error("failed to fetch next segment", "err", err)
		}

		// Send segment back to parent
		sc <- segment

		if segment.ContainsFlag(striping.BlockFlagEndOfData) {
			return
		}

	}

}

func receiveNextSegment(socket DataSocket) (*striping.Segment, error) {
	header := &striping.Header{}
	err := binary.Read(socket, binary.BigEndian, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %s", err)
	}

	if header.IsEODCount() {
		return striping.NewSegmentWithHeader(header, nil), nil
	} else {
		data := make([]byte, header.ByteCount)
		cur := 0

		// Read all bytes
		for {
			n, err := socket.Read(data[cur:header.ByteCount])
			if err != nil {
				return nil, fmt.Errorf("failed to read payload: %s", err)
			}

			cur += n
			if cur == int(header.ByteCount) {
				break
			}
		}

		return striping.NewSegmentWithHeader(header, data), nil
	}

}

var _ io.ReaderFrom = &multisocket{}
var _ io.WriterTo = &multisocket{}

func NewMultisocket(sockets []DataSocket, maxLength int) *multisocket {
	return &multisocket{
		sockets,
		maxLength,
	}
}

func (socket *multisocket) ReadFrom(reader io.Reader) (n int64, err error) {

	segmentChannel := make(chan *striping.Segment)
	parent, child := NewCoordination(len(socket.sockets))

	// Dispatch all sockets
	for _, s := range socket.sockets {
		go writer(s, segmentChannel, child)
	}

	eodc := striping.NewEODCSegment(uint64(len(socket.sockets)))
	segmentChannel <- eodc

	curPos := 0

	for {
		buf := make([]byte, socket.maxLength)
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}

		segmentChannel <- striping.NewSegment(buf[0:n], curPos)
		curPos += n
	}

	// Notify all channels to finish
	// and wait for them
	parent.Stop()

	return

}

func writer(socket DataSocket, sc chan *striping.Segment, coord Child) {
	defer func() {
		eod := striping.NewHeader(0, 0, striping.BlockFlagEndOfData)
		err := sendHeader(socket, eod)
		if err != nil {
			log.Error("Error while sending header", "err", err)
		}
		coord.Done()
	}()

	for {

		select {
		case <-coord.ShouldStop():
			log.Debug("Done", "port", socket.Port())
			return
		case segment := <-sc:
			// log.Debug("New Segment", "hdr", segment.Header)
			err := send(socket, segment)
			if err != nil {
				log.Error("Error while sending packet", "err", err)
			}
		}

	}
}

func sendHeader(socket DataSocket, header *striping.Header) error {
	err := binary.Write(socket, binary.BigEndian, header)
	// log.Debug("Wrote header", "hdr", header)
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
		cur := 0
		for {

			n, err := socket.Write(segment.Data[cur:segment.ByteCount])
			if err != nil {
				return err
			}

			cur += n

			if cur == int(segment.ByteCount) {
				break
			}
		}

	}

	return nil
}

// --------------- Coordination -------------- //

var _ Parent = &coordination{}
var _ Child = &coordination{}

type coordination struct {
	n    int
	stop chan struct{}
	done chan struct{}
}

type Parent interface {
	Stop()
}

type Child interface {
	ShouldStop() chan struct{}
	Done()
}

func NewCoordination(n int) (Parent, Child) {
	c := &coordination{
		n,
		make(chan struct{}),
		make(chan struct{}),
	}
	return Parent(c), Child(c)
}

func (c *coordination) Done() {
	c.done <- struct{}{}
}

func (c *coordination) ShouldStop() chan struct{} {
	return c.stop
}

func (c *coordination) Stop() {
	for i := 0; i < c.n; i++ {
		c.stop <- struct{}{}
	}

	for i := 0; i < c.n; i++ {
		<-c.done
	}
}
