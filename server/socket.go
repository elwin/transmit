// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"

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
}

type ftpActiveSocket struct {
	conn   *net.TCPConn
	host   string
	port   int
	logger Logger
}

func newActiveSocket(remote string, port int, logger Logger, sessionID string) (DataSocket, error) {
	connectTo := net.JoinHostPort(remote, strconv.Itoa(port))

	logger.Print(sessionID, "Opening active data connection to "+connectTo)

	raddr, err := net.ResolveTCPAddr("tcp", connectTo)

	if err != nil {
		logger.Print(sessionID, err)
		return nil, err
	}

	tcpConn, err := net.DialTCP("tcp", nil, raddr)

	if err != nil {
		logger.Print(sessionID, err)
		return nil, err
	}

	socket := new(ftpActiveSocket)
	socket.conn = tcpConn
	socket.host = remote
	socket.port = port
	socket.logger = logger

	return socket, nil
}

func (socket *ftpActiveSocket) Host() string {
	return socket.host
}

func (socket *ftpActiveSocket) Port() int {
	return socket.port
}

func (socket *ftpActiveSocket) Read(p []byte) (n int, err error) {
	return socket.conn.Read(p)
}

func (socket *ftpActiveSocket) ReadFrom(r io.Reader) (int64, error) {
	return socket.conn.ReadFrom(r)
}

func (socket *ftpActiveSocket) Write(p []byte) (n int, err error) {
	return socket.conn.Write(p)
}

func (socket *ftpActiveSocket) Close() error {
	return socket.conn.Close()
}

type ScionSocket struct {
	conn scion.Conn
	port int
}

func (socket ScionSocket) Write(p []byte) (n int, err error) {

	return socket.conn.Write(p)
}

func (socket ScionSocket) Close() error {
	return socket.conn.Close()
}

func (socket ScionSocket) Host() string {
	return socket.conn.LocalAddr().Host.String()
}

func (socket ScionSocket) Read(p []byte) (n int, err error) {

	return socket.conn.Read(p)
}

func (socket ScionSocket) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(socket.conn, r)
}

func (socket ScionSocket) Port() int {

	return socket.port
}

func SendOverSocket(socket DataSocket, header *striping.Header) error {

	return binary.Write(socket, binary.BigEndian, header)

}
