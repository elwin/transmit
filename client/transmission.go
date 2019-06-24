package ftp

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/scion"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"io"
	"sync"
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

	log.Debug(fmt.Sprintf("Received header (%d)", i), "hdr", *header)

	// EOD header, contains no payload
	if header.IsEODCount() {
		transmission.synchronized(func() {
			transmission.eodTotal = header.GetEODCount()
		})

		return finished, nil
	}

	data := make([]byte, header.ByteCount)
	n, err := conn.Read(data)
	if err != nil {
		return false, fmt.Errorf("failed to read payload: %s", err)
	}
	n = n

	log.Debug(fmt.Sprintf("Read %d bytes (%d)", n, i))

	transmission.synchronized(func() {
		transmission.setData(header.OffsetCount, data)
		transmission.eodCount--
	})

	return finished, nil
}

func (transmission *transmission) AcceptData(conns []scion.Conn) error {

	wg := sync.WaitGroup{}

	// i, conn := range conns
	// ^ will lead to errors, don't know why
	for i := range conns {

		wg.Add(1)

		go func(conn scion.Conn, i int) {

			defer wg.Done()

			for {
				finished, err := transmission.ProcessBlock(conn, i)
				if err != nil {
					log.Error("failed to process block", "err", err)
					return
				}

				if finished {
					return
				}
			}

		}(conns[i], i)

	}

	wg.Wait()

	return nil
}

func (transmission *transmission) getData() []byte {
	return transmission.data[:transmission.len]
}
