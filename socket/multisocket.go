package socket

import "io"

type Socket interface {
	io.Reader
	io.Writer
	io.Closer
}

type MultiSocket struct {
	*ReaderSocket
	*WriterSocket
}

var _ DataSocket = MultiSocket{}

func (MultiSocket) Host() string {
	return "my host"
}

func (MultiSocket) Port() int {
	return 69
}

var _ DataSocket = &MultiSocket{}

func NewMultiSocket(sockets []DataSocket) *MultiSocket {
	return &MultiSocket{
		NewReadsocket(sockets),
		NewWriterSocket(sockets, 1000),
	}
}
