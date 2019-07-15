package socket

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/scionproto/scion/go/lib/log"

	"github.com/elwin/transmit/striping"
)

type ReaderSocket struct {
	sockets    []DataSocket
	queue      *striping.SegmentQueue
	written    uint64
	eodc       int
	finished   int // Might need mutex
	dispatched bool
}

var _ io.Reader = &ReaderSocket{}

func NewReadsocket(sockets []DataSocket) *ReaderSocket {
	return &ReaderSocket{
		sockets: sockets,
		queue:   striping.NewSegmentQueue(),
		eodc:    -1,
	}
}

func (s *ReaderSocket) Read(p []byte) (n int, err error) {

	if !s.dispatched {
		s.dispatched = true
		s.dispatchReader()
	}

	// Potential race condition?
	// No, because the reader push segments on the queue
	// before they increase the finished count
	if s.finished == s.eodc && s.queue.Len() == 0 {
		return 0, io.EOF
	}

	for s.queue.Len() == 0 ||
		s.queue.Peek().OffsetCount > s.written {
		// Wait until there is a suitable segment
		time.Sleep(time.Millisecond * 10)
	}

	next := s.queue.Pop()
	s.written += next.ByteCount

	// If copy copies less then the ByteCount we have a problem
	return copy(p, next.Data), nil

}

func (s *ReaderSocket) dispatchReader() {
	for _, subSocket := range s.sockets {
		go s.receiveOnSocket(subSocket)
	}
}

func (s *ReaderSocket) receiveOnSocket(socket DataSocket) {
	for {

		seg, err := receiveNextSegment(socket)
		if err != nil {
			log.Error("Failed to receive segment", "err", err)
		}

		// The EOD count header has a special format
		// and is only used to transmit the EOD count
		if seg.IsEODCount() {
			s.eodc = seg.GetEODCount()
			continue
		}

		s.queue.Push(seg)

		if seg.ContainsFlag(striping.BlockFlagEndOfData) {
			s.finished++
		}

		if seg.ContainsFlag(striping.BlockFlagSenderClosesConnection) {
			log.Debug("Closing connection")
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
				return striping.NewSegmentWithHeader(header, data), nil
			}
		}

	}

}
