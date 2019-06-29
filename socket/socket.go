package socket

import (
	"encoding/binary"
	"io"

	"github.com/elwin/transmit/scion"
	"github.com/elwin/transmit/striping"
)

// DataSocket describes a data parallelSockets is used to send non-control data between the client and
// server.
type DataSocket interface {
	Host() string

	Port() int

	// the standard io.Reader interface
	Read(p []byte) (n int, err error)

	// the standard io.ReaderFrom interface
	ReadFrom(r io.Reader) (int64, error)

	// the standard io.Writer interface
	Write(p []byte) (n int, err error)

	// the standard io.Closer interface
	Close() error

	SendHeader(h *striping.Header) error
}

var _ DataSocket = &ScionSocket{}

type ScionSocket struct {
	conn scion.Conn
	port int
}

func NewScionSocket(conn scion.Conn, port int) *ScionSocket {
	return &ScionSocket{conn, port}
}

func (socket *ScionSocket) Write(p []byte) (n int, err error) {

	return socket.conn.Write(p)
}

func (socket *ScionSocket) Close() error {
	return socket.conn.Close()
}

func (socket *ScionSocket) Host() string {
	return socket.conn.LocalAddr().Host.String()
}

func (socket *ScionSocket) Read(p []byte) (n int, err error) {

	return socket.conn.Read(p)
}

func (socket *ScionSocket) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(socket.conn, r)
}

func (socket *ScionSocket) Port() int {

	return socket.port
}

func (socket *ScionSocket) SendHeader(h *striping.Header) error {
	return binary.Write(socket, binary.BigEndian, h)
}
