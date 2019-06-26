package striping

type Segment struct {
	*Header
	Data []byte
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
