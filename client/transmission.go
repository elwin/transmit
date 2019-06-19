package ftp

import (
	"encoding/binary"
	"fmt"
	"github.com/elwin/transmit/scion"
	"github.com/elwin/transmit/striping"
	"github.com/scionproto/scion/go/lib/log"
	"strconv"
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
	return &transmission{data: make([]byte, 2)}
}

func (transmission *transmission) setData(offset uint64, data []byte) {
	required := offset + uint64(len(data))
	actual := uint64(len(transmission.data))

	// Grow capacity if necessary
	if required > actual {

		if required < 2*actual {
			required = 2 * actual
		}

		fmt.Printf("Required: %d - Actual: %d\n", required, actual)

		temp := make([]byte, required)
		copy(temp, transmission.data)
		transmission.data = temp

		required = offset + uint64(len(data))
		actual = uint64(len(transmission.data))
	}

	copy(transmission.data[offset:], data)
	fmt.Println(transmission.data)
}

func (transmission *transmission) Completed() bool {
	return transmission.eodTotal > 0 &&
		uint64(transmission.eodCount) >= transmission.eodTotal
}

func (transmission *transmission) AcceptData(conns []scion.Conn) error {

	wg := sync.WaitGroup{}

	for i, conn := range conns {

		wg.Add(1)

		func(addr string) {
			defer wg.Done()

			for {
				header := &striping.Header{}
				err := binary.Read(conn, binary.BigEndian, header)
				if err != nil {
					log.Error("failed to read header", "err", err)
				}

				log.Debug(fmt.Sprintf("Read header on %s", addr), "hdr", *header)

				// EOD header, contains no data
				if header.IsEODCount() {
					transmission.Lock()
					transmission.eodTotal = header.GetEODCount()
					transmission.Unlock()
					continue
				}

				data := make([]byte, header.ByteCount)
				n, err := conn.Read(data)
				if err != nil {
					log.Error("failed to read data", "err", err)
				}

				log.Debug(fmt.Sprintf("Read %d bytes on %s", n, addr))
				transmission.Lock()
				transmission.setData(header.OffsetCount, data)
				transmission.eodCount--
				transmission.Unlock()

				if header.ContainsFlag(striping.BlockFlagEndOfData) {
					fmt.Println("EOD")
					return
				}
			}

		}(string(strconv.Itoa(i)))

	}

	wg.Wait()
	return nil
}
