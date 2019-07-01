package striping

import "github.com/elwin/transmit/queue"

type Segment struct {
	*Header
	Data []byte
}

func (a *Segment) Less(b queue.Sortable) bool {
	return a.OffsetCount < b.(*Segment).OffsetCount
}

func NewSegment(data []byte, offset int, flags ...uint8) *Segment {

	return &Segment{
		NewHeader(uint64(len(data)), uint64(offset), flags...),
		data,
	}

}

func NewEODCSegment(count uint64) *Segment {
	return &Segment{NewEODCHeader(count), nil}
}

func NewSegmentWithHeader(header *Header, data []byte) *Segment {
	return &Segment{
		header,
		data,
	}
}
