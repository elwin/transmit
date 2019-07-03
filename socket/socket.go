package socket

import (
	"io"

	"github.com/elwin/transmit/scion"
)

// DataSocket describes a data socket used to send non-control data between the client and
// server.
type DataSocket interface {
	Host() string
	Port() int

	io.Reader
	io.ReaderFrom
	io.Writer
	io.Closer
}

var _ DataSocket = &ScionSocket{}

type ScionSocket struct {
	conn scion.Conn
	port int
}

func (socket *ScionSocket) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, socket.conn)
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
