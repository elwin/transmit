package ftp

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/striping"
	"io"
)

// Requires synchronisation
type Transmission struct {
	eodTotal uint64
	eodCount uint64
	data     []byte
}

func (transmission Transmission) setData(offset uint64, data []byte) {

	// Grow capacity if necessary
	if offset+uint64(len(data)) > uint64(len(transmission.data)) {
		temp := make([]byte, len(transmission.data), (cap(transmission.data)+1)*2)
		copy(temp, transmission.data)
		transmission.data = temp
	}

	data[offset:] = data

}

func (transmission Transmission) Read(reader io.Reader) (finished bool, err error) {

	var header striping.Header
	err := binary.Read(reader, binary.BigEndian, &header)

	if err != nil {
		return false, fmt.Errorf("failed to parse header: %s", err)
	}

	if header.IsEODCount() {
		transmission.eodTotal = header.GetEODCount()
		return transmission.eodTotal <= transmission.eodCount, nil
	}

	var data = make([]byte, header.ByteCount)
	n, err := reader.Read(data)
	if err != nil {
		return false, fmt.Errorf("failed to read data: %s", err)
	}

	if uint64(n) < header.ByteCount {
		return false, fmt.Errorf("failed to read enoguh data, expected %d but got %d", header.ByteCount, n)
	}

	transmission.setData(header.OffsetCount, data)
	transmission.eodCount--

	finished = transmission.eodTotal > 0 &&
		transmission.eodTotal <= transmission.eodCount

	return finished, nil
}
