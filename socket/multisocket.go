package socket

import "io"

type Socket interface {
	io.Reader
	io.Writer
	io.Closer
}

type MutliSocket struct {
	io.Reader
	io.WriteCloser
}

/*
func NewMultiSocket(sockets []DataSocket) *MutliSocket {
	return &MutliSocket{
		NewReadsocket(sockets),
		NewWriterSocket(sockets),
	}
}
*/
