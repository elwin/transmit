package ftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/elwin/transmit/scion"
	"github.com/elwin/transmit/striping"
)

// Fix uint / int difference
type transmission struct {
	eodTotal uint64
	eodCount int64
	len      uint64
	data     []byte
	sync.Mutex
}

func NewTransmission() *transmission {
	return &transmission{data: make([]byte, 2)}
}

func (transmission *transmission) synchronized(f func()) {
	transmission.Lock()
	defer transmission.Unlock()

	f()
}

func (transmission *transmission) setData(offset uint64, data []byte) {
	required := offset + uint64(len(data))
	actual := uint64(len(transmission.data))

	// Grow capacity if necessary
	if required > actual {

		if required < 2*actual {
			required = 2 * actual
		}

		temp := make([]byte, required)
		copy(temp, transmission.data)
		transmission.data = temp

		required = offset + uint64(len(data))
		actual = uint64(len(transmission.data))
	}

	if required > transmission.len {
		transmission.len = required
	}

	copy(transmission.data[offset:], data)
}

func (transmission *transmission) Completed() bool {
	return transmission.eodTotal > 0 &&
		uint64(transmission.eodCount) >= transmission.eodTotal
}

func (transmission *transmission) ProcessBlock(conn io.Reader, i int) (finished bool, err error) {

	header := &striping.Header{}
	err = binary.Read(conn, binary.BigEndian, header)
	if err != nil {
		return false, fmt.Errorf("failed to read header: %s", err)
	}

	finished = header.ContainsFlag(striping.BlockFlagEndOfData)

	// log.Debug(fmt.Sprintf("Received header (%d)", i), "hdr", *header)

	// EOD header, contains no payload
	if header.IsEODCount() {
		transmission.synchronized(func() {
			transmission.eodTotal = uint64(header.GetEODCount())
		})

		return finished, nil
	}

	data := make([]byte, header.ByteCount)
	cur := 0

	// Read all bytes
	for {
		n, err := conn.Read(data[cur:header.ByteCount])
		if err != nil {
			return false, fmt.Errorf("failed to read payload: %s", err)
		}

		cur += n
		if cur == int(header.ByteCount) {
			break
		}
	}

	// log.Debug(fmt.Sprintf("Read %d bytes (%d)", n, i))

	transmission.synchronized(func() {
		transmission.setData(header.OffsetCount, data)
		transmission.eodCount--
	})

	return finished, nil
}

func (transmission *transmission) AcceptData(conns []scion.Conn) error {

	return nil
}

func (transmission *transmission) getData() []byte {
	return transmission.data[:transmission.len]
}
