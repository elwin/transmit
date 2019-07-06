package socket

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"io"
)

type WriterSocket struct {
	sockets        []DataSocket
	maxLength      int
	segmentChannel chan *striping.Segment
	parent         Parent
	child          Child
	written        int
	writing        bool
}

var _ io.ReaderFrom = &WriterSocket{}
var _ io.Writer = &WriterSocket{}
var _ io.Closer = &WriterSocket{}

func NewWriterSocket(sockets []DataSocket, maxLength int) *WriterSocket {
	parent, child := NewCoordination(len(sockets))

	return &WriterSocket{
		sockets:        sockets,
		maxLength:      maxLength,
		segmentChannel: make(chan *striping.Segment),
		parent:         parent,
		child:          child,
	}
}

func (s *WriterSocket) StreamWriter() {

	// Dispatch all sockets
	for _, socket := range s.sockets {
		go s.writer(socket)
	}

	eodc := striping.NewEODCSegment(uint64(len(s.sockets)))
	s.segmentChannel <- eodc

}

func (s *WriterSocket) Write(p []byte) (n int, err error) {
	if !s.writing {
		go s.StreamWriter()
	}

	length := s.maxLength
	if length > len(p) {
		length = len(p)
	}

	s.segmentChannel <- striping.NewSegment(p[0:length], s.written)
	s.written += length

	return length, nil
}

func (s *WriterSocket) ReadFrom(reader io.Reader) (n int64, err error) {
	if !s.writing {
		go s.StreamWriter()
	}

	for {
		buf := make([]byte, s.maxLength)
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}

		s.segmentChannel <- striping.NewSegment(buf[0:n], s.written)
		s.written += n
	}

	// Notify all channels to finish
	// and wait for them
	s.parent.Stop()

	return

}

func (s *WriterSocket) writer(socket DataSocket) {
	defer func() {
		eod := striping.NewHeader(0, 0, striping.BlockFlagEndOfData)
		err := sendHeader(socket, eod)
		if err != nil {
			log.Error("Error while sending header", "err", err)
		}
		s.child.Done()
	}()

	for {

		select {
		case <-s.child.ShouldStop():
			log.Debug("Done", "port", socket.Port())
			return
		case segment := <-s.segmentChannel:
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

func (s *WriterSocket) Close() error {
	// TODO
	// First Flush / Block Channels
	// Then send closing message
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
