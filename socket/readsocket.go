package socket

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"io"
)

type readsocket struct {
	sockets  []DataSocket
	queue    *striping.SegmentQueue
	written  uint64
	eodc     int
	finished int
}

var _ io.Reader = &readsocket{}

func (s *readsocket) Read(p []byte) (n int, err error) {
	if s.finished == s.eodc && s.queue.Len() == 0 {
		return 0, io.EOF
	}

	for s.queue.Len() == 0 ||
		s.queue.Peek().OffsetCount < s.written {
		// Loop and wait until at least one available element is here
	}

	next := s.queue.Pop()
	s.written += next.ByteCount

	if len(p) < len(next.Data) {
	}

	return copy(p, next.Data), nil
}

func NewReadsocket(sockets []DataSocket) *readsocket {
	s := &readsocket{
		sockets,
		striping.NewSegmentQueue(),
		0,
		-1,
		0,
	}

	go s.writeToQueue()

	return s
}

func (s *readsocket) writeToQueue() {

	segmentChannel := make(chan *striping.Segment)
	done := make(chan struct{})

	for _, s := range s.sockets {
		go reader(s, segmentChannel, done)
	}

	for {

		// Listen to children
		select {
		case segment := <-segmentChannel:
			if segment.IsEODCount() {
				s.eodc = segment.GetEODCount()
			} else {
				s.queue.Push(segment)
			}

		case <-done:
			s.finished += 1
			if s.finished == s.eodc {
				fmt.Println("Finished")
				return
			}
		}
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
