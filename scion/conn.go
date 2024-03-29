package scion

import (
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/snet"
)

// Copied from the net package and replaced net.Addr with snet.Addr
type Conn interface {
	// Read reads data from the connection.
	// Read can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	Read(b []byte) (n int, err error)

	// ReadFrom writes data to the connection.
	// ReadFrom can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	// Any blocked Read or ReadFrom operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the Local network Local.
	LocalAddr() snet.Addr

	// RemoteAddr returns the Local network Local.
	RemoteAddr() snet.Addr

	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	//
	// A deadline is an absolute time after which I/O operations
	// fail with a timeout (see type Error) instead of
	// blocking. The deadline applies to all future and pending
	// I/O, not just the immediately following call to Read or
	// ReadFrom. After a deadline has been exceeded, the connection
	// can be refreshed by setting a deadline in the future.
	//
	// An idle timeout can be implemented by repeatedly extending
	// the deadline after successful Read or ReadFrom calls.
	//
	// A zero value for t means I/O operations will not time out.
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future ReadFrom calls
	// and any currently-blocked ReadFrom call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means ReadFrom will not time out.
	SetWriteDeadline(t time.Time) error
}

var _ Conn = &connection{}

type connection struct {
	quic.Stream
	Local  snet.Addr
	Remote snet.Addr
}

func NewConnection(stream quic.Stream, local, remote snet.Addr) *connection {
	return &connection{
		stream,
		local,
		remote,
	}
}

func (connection *connection) Read(b []byte) (n int, err error) {
	return connection.Stream.Read(b)
}

func (connection *connection) Write(b []byte) (n int, err error) {
	return connection.Stream.Write(b)
}

func (connection *connection) LocalAddr() snet.Addr {
	return connection.Local
}

func (connection *connection) RemoteAddr() snet.Addr {
	return connection.Remote
}

func (connection *connection) Close() error {
	return connection.Stream.Close()
}
