package striping

// Extended Block Header Flags
const (
	BlockFlagEndOfRecord            uint8 = 128 // Legacy
	BlockFlagEndOfDataCount         uint8 = 64
	BlockFlagSuspectErrors          uint8 = 32
	BlockFlagRestartMarker          uint8 = 16 // Legacy
	BlockFlagEndOfData              uint8 = 8
	BlockFlagSenderClosesConnection uint8 = 4
)

type Header struct {
	Descriptor  uint8
	ByteCount   uint64
	OffsetCount uint64
}

func NewHeader(byteCount, OffsetCount uint64, flags ...uint8) Header {
	header := Header{
		ByteCount:   byteCount,
		OffsetCount: OffsetCount,
	}
	for _, flag := range flags {
		header.Descriptor |= flag
	}

	return header
}

func (header Header) ContainsFlag(flag uint8) bool {
	return header.Descriptor&flag == flag
}
