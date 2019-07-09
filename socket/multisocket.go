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
	return "hostaddress"
}

func (MultiSocket) Port() int {
	return 0
}

var _ DataSocket = &MultiSocket{}

func NewMultiSocket(sockets []DataSocket, maxLength int) *MultiSocket {
	return &MultiSocket{
		NewReadsocket(sockets),
		NewWriterSocket(sockets, maxLength),
	}
}
