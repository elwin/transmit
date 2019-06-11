package ftp

// FTP modes
const (
	ModeStream            = 'S'
	ModeExtendedBlockMode = 'E'
)

// Extended Block Header Flags
const (
	BlockFlagEndOfRecord            = 128 // Legacy
	BlockFlagEndOfDataCount         = 64
	BlockFlagSuspectErrors          = 32
	BlockFlagRestartMarker          = 16 // Legacy
	BlockFlagEndOfData              = 8
	BlockFlagSenderClosesConnection = 4
)

func AssembleFlags(flags ...int) (header int) {

	header = 0
	for _, flag := range flags {
		header = header | flag
	}

	return
}

func ContainsFlag(header, flag int) bool {
	return header&flag == flag
}

const (
	PartialFileTransport = "PFT"
)
