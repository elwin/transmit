package striping

// Extended Block Header Flags
const (
	// Deprecated: Around for legacy purposes
	BlockFlagEndOfRecord    uint8 = 128
	BlockFlagEndOfDataCount uint8 = 64
	BlockFlagSuspectErrors  uint8 = 32
	// Deprecated: Around for legacy purposes
	BlockFlagRestartMarker          uint8 = 16
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

	return header.AddFlag(flags...)
}

func NewEODCHeader(eodCount uint64, flags ...uint8) Header {
	return NewHeader(0, eodCount, append(flags, BlockFlagEndOfDataCount)...)
}

func (header Header) ContainsFlag(flag uint8) bool {
	return header.Descriptor&flag == flag
}

func (header Header) AddFlag(flags ...uint8) Header {
	for _, flag := range flags {
		header.Descriptor |= flag
	}

	return header
}

func (header Header) GetEODCount() uint64 {
	return header.OffsetCount
}

func (header Header) IsEODCount() bool {
	return header.ContainsFlag(BlockFlagEndOfDataCount)
}
