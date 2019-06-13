// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/elwin/transmit/scion"
	"github.com/elwin/transmit/striping"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
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

type ftpPassiveSocket struct {
	conn      net.Conn
	port      int
	host      string
	ingress   chan []byte
	egress    chan []byte
	logger    Logger
	lock      sync.Mutex // protects conn and err
	err       error
	tlsConfig *tls.Config
}

// Detect if an error is "bind: address already in use"
//
// Originally from https://stackoverflow.com/a/52152912/164234
func isErrorAddressAlreadyInUse(err error) bool {
	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}

/*
func newPassiveSocket(host string, port func() int, logger Logger, sessionID string, tlsConfig *tls.Config) (DataSocket, error) {
	/*
		parallelSockets := new(ftpPassiveSocket)
		parallelSockets.ingress = make(chan []byte)
		parallelSockets.egress = make(chan []byte)
		parallelSockets.logger = logger
		parallelSockets.host = host
		parallelSockets.tlsConfig = tlsConfig
		const retries = 10
		var err error
		for i := 1; i <= retries; i++ {
			parallelSockets.port = port()
			err = parallelSockets.GoListenAndServe(sessionID)
			if err != nil && parallelSockets.port != 0 && isErrorAddressAlreadyInUse(err) {
				// choose a different port on error already in use
				continue
			}
			break
		}
		return parallelSockets, err
*/
/*

	fmt.Println("Trying to create a parallelSockets")

	listener, err := scion.Listen("1-ff00:0:110,[127.0.0.1]:40000")

	if err != nil {
		return nil, err
	}

	fmt.Println("Listening")

	stream, err := listener.Accept()

	fmt.Println("Accepted")

	return ScionSocket{stream, 40002}, err
}

func (parallelSockets *ftpPassiveSocket) Host() string {
	return parallelSockets.host
}

func (parallelSockets *ftpPassiveSocket) Port() int {
	return parallelSockets.port
}

func (parallelSockets *ftpPassiveSocket) Read(p []byte) (n int, err error) {
	parallelSockets.lock.Lock()
	defer parallelSockets.lock.Unlock()
	if parallelSockets.err != nil {
		return 0, parallelSockets.err
	}
	return parallelSockets.conn.Read(p)
}

func (parallelSockets *ftpPassiveSocket) ReadFrom(r io.Reader) (int64, error) {
	parallelSockets.lock.Lock()
	defer parallelSockets.lock.Unlock()
	if parallelSockets.err != nil {
		return 0, parallelSockets.err
	}

	// For normal TCPConn, this will use sendfile syscall; if not,
	// it will just downgrade to normal read/write procedure
	return io.Copy(parallelSockets.conn, r)
}

*/

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

func SendOverSocket(socket DataSocket, header striping.Header) error {
	return binary.Write(socket, binary.BigEndian, header)

}

/*

func (parallelSockets *ftpPassiveSocket) Write(p []byte) (n int, err error) {
	parallelSockets.lock.Lock()
	defer parallelSockets.lock.Unlock()
	if parallelSockets.err != nil {
		return 0, parallelSockets.err
	}
	return parallelSockets.conn.Write(p)
}

func (parallelSockets *ftpPassiveSocket) Close() error {
	parallelSockets.lock.Lock()
	defer parallelSockets.lock.Unlock()
	if parallelSockets.conn != nil {
		return parallelSockets.conn.Close()
	}
	return nil
}

func (parallelSockets *ftpPassiveSocket) GoListenAndServe(sessionID string) (err error) {
	laddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("", strconv.Itoa(parallelSockets.port)))
	if err != nil {
		parallelSockets.logger.Print(sessionID, err)
		return
	}

	var tcplistener *net.TCPListener
	tcplistener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		parallelSockets.logger.Print(sessionID, err)
		return
	}

	// The timeout, for a remote client to establish connection
	// with a PASV style data connection.
	const acceptTimeout = 60 * time.Second
	err = tcplistener.SetDeadline(time.Now().Add(acceptTimeout))
	if err != nil {
		parallelSockets.logger.Print(sessionID, err)
		return
	}

	var listener net.Listener = tcplistener
	add := listener.Addr()
	parts := strings.Split(add.String(), ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		parallelSockets.logger.Print(sessionID, err)
		return
	}

	parallelSockets.port = port
	if parallelSockets.tlsConfig != nil {
		listener = tls.NewListener(listener, parallelSockets.tlsConfig)
	}

	parallelSockets.lock.Lock()
	go func() {
		defer parallelSockets.lock.Unlock()

		conn, err := listener.Accept()
		if err != nil {
			parallelSockets.err = err
			return
		}
		parallelSockets.err = nil
		parallelSockets.conn = conn
		_ = listener.Close()
	}()
	return nil
}

*/
