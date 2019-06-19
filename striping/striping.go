package striping

// Extended Block Header Flags
const (
	BlockFlagEndOfDataCount         uint8 = 64
	BlockFlagSuspectErrors          uint8 = 32
	BlockFlagEndOfData              uint8 = 8
	BlockFlagSenderClosesConnection uint8 = 4

	// Deprecated: Around for legacy purposes
	BlockFlagEndOfRecord uint8 = 128
	// Deprecated: Around for legacy purposes
	BlockFlagRestartMarker uint8 = 16
)

type Header struct {
	Descriptor  uint8
	ByteCount   uint64
	OffsetCount uint64
}

func NewHeader(byteCount, offsetCount uint64, flags ...uint8) *Header {
	header := Header{
		ByteCount:   byteCount,
		OffsetCount: offsetCount,
	}

	header.AddFlag(flags...)

	return &header
}

func NewEODCHeader(eodCount uint64, flags ...uint8) *Header {
	return NewHeader(
		0,
		eodCount,
		append(flags, BlockFlagEndOfDataCount)...)
}

func (header *Header) ContainsFlag(flag uint8) bool {
	return header.Descriptor&flag == flag
}

func (header *Header) AddFlag(flags ...uint8) {
	for _, flag := range flags {
		header.Descriptor |= flag
	}
}

func (header *Header) GetEODCount() uint64 {
	return header.OffsetCount
}

func (header *Header) IsEODCount() bool {
	return header.ContainsFlag(BlockFlagEndOfDataCount)
}
