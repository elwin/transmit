package ftp

import (
	"github.com/elwin/transmit/socket"
	"io"
	"time"
)

// Response represents a data-connection
type Response interface {
	io.Reader
	SetDeadline(time time.Time) error
}

var _ Response = &ConnResponse{}

// Response represents a data-connection
type ConnResponse struct {
	conn   socket.DataSocket
	c      *ServerConn
	closed bool
}

// Read implements the io.Reader interface on a FTP data connection.
func (r *ConnResponse) Read(buf []byte) (int, error) {
	return r.conn.Read(buf)
}

// Close implements the io.Closer interface on a FTP data connection.
// After the first call, Close will do nothing and return nil.
/*
func (r *ConnResponse) Close() error {

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
}*/

// SetDeadline sets the deadlines associated with the connection.
func (r *ConnResponse) SetDeadline(t time.Time) error {
	// return r.conn.SetDeadline(t)
	return nil
}
