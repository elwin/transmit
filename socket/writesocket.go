package socket

import (
	"encoding/binary"
	"io"

	"github.com/scionproto/scion/go/lib/log"

	"github.com/elwin/transmit/striping"
)

type WriterSocket struct {
	sockets           []DataSocket
	maxLength         int
	segmentChannel    chan *striping.Segment
	parent            Parent
	child             Child
	written           int
	dispatchedWriters bool
}

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

func (s *WriterSocket) Write(p []byte) (n int, err error) {
	if !s.dispatchedWriters {
		s.dispatchedWriters = true
		s.dispatchWriter()
	}

	cur := 0

	for {
		if cur == len(p) {
			return cur, nil
		}

		to := cur + s.maxLength
		if to > len(p) {
			to = len(p)
		}

		data := make([]byte, to-cur)
		copy(data, p[cur:to])

		s.segmentChannel <- striping.NewSegment(data, s.written)

		s.written += cur - to
		cur = to
	}
}

func (s *WriterSocket) dispatchWriter() {
	for _, socket := range s.sockets {
		go s.writer(socket)
	}

	eodc := striping.NewEODCSegment(uint64(len(s.sockets)))
	s.segmentChannel <- eodc
}

func (s *WriterSocket) writer(socket DataSocket) {
	for {
		select {
		case segment := <-s.segmentChannel:
			err := writeSegment(segment, socket)
			if err != nil {
				log.Error("Failed to write segment", "err", err)
			}

		case <-s.child.ShouldStop():
			eod := striping.NewHeader(
				0, 0,
				striping.BlockFlagEndOfData)
			err := writeHeader(eod, socket)
			if err != nil {
				log.Error("Failed to write eod header", "err", err)
			}
			s.child.Done()
			return
		}
	}
}

// Closing the WriterSocket blocks until until all
// children have finished senidng and then sends the
// closing connection header over all sub-sockets
// to signal closing the connection
func (s *WriterSocket) Close() error {

	// Wait until all sockets finished sending
	s.parent.Wait()

	for i, subSocket := range s.sockets {
		closingHeader := striping.NewClosingHeader()
		err := writeHeader(closingHeader, subSocket)
		if err != nil {
			log.Debug("Failed to write header", "err", err)
		}
		s.sockets[i] = nil
	}

	return nil
}

// Helper Functions

func writeHeader(header *striping.Header, socket DataSocket) error {
	return binary.Write(socket, binary.BigEndian, header)
}

func writeSegment(segment *striping.Segment, socket DataSocket) error {
	err := writeHeader(segment.Header, socket)
	if err != nil {
		return err
	}

	if segment.IsEODCount() {
		return nil
	}

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

	return nil
}
