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

// Requires synchronisation
// Fix uint / int difference
type transmission struct {
	eodTotal uint64
	eodCount int64
	data     []byte
	sync.Mutex
}

func NewTransmission() *transmission {
	return &transmission{data: make([]byte, 10)}
}

func (transmission *transmission) setData(offset uint64, data []byte) {
	required := offset + uint64(len(data))
	actual := uint64(cap(transmission.data))

	// Grow capacity if necessary
	for required > actual {
		temp := make([]byte, len(transmission.data), (cap(transmission.data)+1)*2)
		copy(temp, transmission.data)
		transmission.data = temp

		required = offset + uint64(len(data))
		actual = uint64(cap(transmission.data))
	}

	copy(transmission.data[offset:], data)
}

func (transmission *transmission) Completed() bool {
	return transmission.eodTotal > 0 && transmission.eodCount >= transmission.eodCount
}

func (transmission *transmission) Read(reader io.Reader) (finished bool, err error) {

	var header striping.Header
	err = binary.Read(reader, binary.BigEndian, &header)
	finished = header.ContainsFlag(striping.BlockFlagEndOfData)

	log.Debug("Received header", "hdr", header)

	if err != nil {
		return true, fmt.Errorf("failed to parse header: %s", err)
	}

	if header.IsEODCount() {
		log.Debug("Received EDOC", "count", header.GetEODCount())
		transmission.Lock()
		defer transmission.Unlock()

		transmission.eodTotal = header.GetEODCount()
		return finished, nil
	}

	var data = make([]byte, header.ByteCount)
	n, err := reader.Read(data)
	if err != nil {
		return finished, fmt.Errorf("failed to read data: %s", err)
	}

	if uint64(n) < header.ByteCount {
		return finished, fmt.Errorf("failed to read enough data, expected %d but got %d", header.ByteCount, n)
	}

	transmission.Lock()
	defer transmission.Unlock()

	transmission.setData(header.OffsetCount, data)
	transmission.eodCount--

	return finished, nil
}

func (transmission *transmission) AcceptData(conns []scion.Conn) error {

	wg := sync.WaitGroup{}

	for _, conn := range conns {

		wg.Add(1)

		func() {
			defer wg.Done()

			for {
				finished, err := transmission.Read(conn)
				if err != nil {
					log.Error("Something happened when trying to write", "err", err)
					return
				} else {
					log.Debug("Received some part", "data", transmission.data)
				}

				if finished {
					return
				}
			}
		}()

	}

	wg.Wait()

	return nil
}
