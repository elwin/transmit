package ftp

import (
	"github.com/elwin/transmit/socket"
	"io"
	"time"
)

// Response represents a data-connection
type Response interface {
	io.ReadCloser
	SetDeadline(time time.Time) error
}

var _ Response = &SingleConnectionResponse{}

// Response represents a data-connection
type SingleConnectionResponse struct {
	conn   socket.DataSocket
	c      *ServerConn
	closed bool
}

// Read implements the io.Reader interface on a FTP data connection.
func (r *SingleConnectionResponse) Read(buf []byte) (int, error) {
	return r.conn.Read(buf)
}

// Close implements the io.Closer interface on a FTP data connection.
// After the first call, Close will do nothing and return nil.
func (r *SingleConnectionResponse) Close() error {
	if r.closed {
		return nil
	}
	err := r.conn.Close()
	_, _, err2 := r.c.conn.ReadResponse(StatusClosingDataConnection)
	if err2 != nil {
		err = err2
	}
	r.closed = true
	return err
}

// SetDeadline sets the deadlines associated with the connection.
func (r *SingleConnectionResponse) SetDeadline(t time.Time) error {
	// return r.conn.SetDeadline(t)
	return nil
}

var _ Response = &MultiConnectionResponse{}

type MultiConnectionResponse struct {
	io.Reader
}

func (MultiConnectionResponse) Close() error {
	return nil
}

// What are we supposed to do?
func (MultiConnectionResponse) SetDeadline(time time.Time) error {
	return nil
}
