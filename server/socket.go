// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"io"
	"net"
	"strconv"

	"github.com/elwin/transmit/striping"

	"github.com/elwin/transmit/socket"
)

var _ socket.DataSocket = &ftpActiveSocket{}

type ftpActiveSocket struct {
	conn   *net.TCPConn
	host   string
	port   int
	logger Logger
}

func (socket *ftpActiveSocket) SendHeader(h *striping.Header) error {
	panic("implement me")
}

func newActiveSocket(remote string, port int, logger Logger, sessionID string) (socket.DataSocket, error) {
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
