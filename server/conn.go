// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	random "math/rand"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/elwin/transmit/socket"

	"github.com/elwin/transmit/scion"
)

const (
	defaultWelcomeMessage = "Welcome to the Go FTP Server"
)

type Conn struct {
	conn            scion.Conn
	controlReader   *bufio.Reader
	controlWriter   *bufio.Writer
	socket          socket.DataSocket
	parallelSockets *socket.MultiSocket
	driver          Driver
	auth            Auth
	logger          Logger
	server          *Server
	tlsConfig       *tls.Config
	sessionID       string
	namePrefix      string
	reqUser         string
	user            string
	renameFrom      string
	lastFilePos     int64
	appendData      bool
	closed          bool
	tls             bool
	extendedMode    bool
}

func (conn *Conn) LoginUser() string {
	return conn.user
}

func (conn *Conn) IsLogin() bool {
	return len(conn.user) > 0
}

func (conn *Conn) PublicIp() string {
	return conn.server.PublicIp
}

func (conn *Conn) passiveListenIP() string {
	if len(conn.PublicIp()) > 0 {
		return conn.PublicIp()
	}
	return conn.conn.LocalAddr().Host.String()
}

// TODO: Replace generation with proper function
func (conn *Conn) PassivePort() int {

	return random.Intn(10000) + 40000
}

// returns a random 20 char string that can be used as a unique session ID
func newSessionID() string {
	hash := sha256.New()
	_, err := io.CopyN(hash, rand.Reader, 50)
	if err != nil {
		return "????????????????????"
	}
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[0:20]
}

// Serve starts an endless loop that reads FTP commands from the client and
// responds appropriately. terminated is a channel that will receive a true
// message when the connection closes. This loop will be running inside a
// goroutine, so use this channel to be notified when the connection can be
// cleaned up.
func (conn *Conn) Serve() {
	conn.logger.Print(conn.sessionID, "connection Established")
	// send welcome
	conn.writeMessage(220, conn.server.WelcomeMessage)
	// read commands
	for {
		line, err := conn.controlReader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				conn.logger.Print(conn.sessionID, fmt.Sprint("read error:", err))
			}

			break
		}
		conn.receiveLine(line)
		// QUIT command closes connection, break to avoid error on reading from
		// closed sockets
		if conn.closed == true {
			break
		}
	}
	conn.Close()
	conn.logger.Print(conn.sessionID, "connection Terminated")
}

// Close will manually close this connection, even if the client isn't ready.
func (conn *Conn) Close() {
	conn.conn.Close()
	conn.closed = true
	if conn.socket != nil {
		conn.socket.Close()
		conn.socket = nil
	}

	if conn.parallelSockets != nil {
		conn.parallelSockets.Close()
		conn.parallelSockets = nil
	}
}

func (conn *Conn) upgradeToTLS() error {

	return nil
	/*

		conn.logger.Print(conn.sessionID, "Upgrading connectiion to TLS")
		tlsConn := tls.Server(conn.conn, conn.tlsConfig)
		err := tlsConn.Handshake()
		if err == nil {
			conn.conn = tlsConn
			conn.controlReader = bufio.NewReader(tlsConn)
			conn.controlWriter = bufio.NewWriter(tlsConn)
			conn.tls = true
		}
		return err
	*/
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (conn *Conn) receiveLine(line string) {
	command, param := conn.parseLine(line)
	conn.logger.PrintCommand(conn.sessionID, command, param)
	cmdObj := commands[strings.ToUpper(command)]
	if cmdObj == nil {
		conn.writeMessage(500, "Command not found")
		return
	}
	if cmdObj.RequireParam() && param == "" {
		conn.writeMessage(553, "action aborted, required param missing")
	} else if cmdObj.RequireAuth() && conn.user == "" {
		conn.writeMessage(530, "not logged in")
	} else {
		cmdObj.Execute(conn, param)
	}
}

func (conn *Conn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

// writeMessage will send a standard FTP response back to the client.
func (conn *Conn) writeMessage(code int, message string) (wrote int, err error) {
	conn.logger.PrintResponse(conn.sessionID, code, message)
	line := fmt.Sprintf("%d %s\r\n", code, message)
	wrote, err = conn.controlWriter.WriteString(line)
	conn.controlWriter.Flush()
	return
}

// writeMessage will send a standard FTP response back to the client.
func (conn *Conn) writeMessageMultiline(code int, message string) (wrote int, err error) {
	conn.logger.PrintResponse(conn.sessionID, code, message)
	line := fmt.Sprintf("%d-%s\r\n%d END\r\n", code, message, code)
	wrote, err = conn.controlWriter.WriteString(line)
	conn.controlWriter.Flush()
	return
}

// buildPath takes a client supplied path or filename and generates a safe
// absolute path within their account sandbox.
//
//    buildpath("/")
//    => "/"
//    buildpath("one.txt")
//    => "/one.txt"
//    buildpath("/files/two.txt")
//    => "/files/two.txt"
//    buildpath("files/two.txt")
//    => "/files/two.txt"
//    buildpath("/../../../../etc/passwd")
//    => "/etc/passwd"
//
// The driver implementation is responsible for deciding how to treat this path.
// Obviously they MUST NOT just read the path off disk. The probably want to
// prefix the path with something to scope the users access to a sandbox.
func (conn *Conn) buildPath(filename string) (fullPath string) {
	if len(filename) > 0 && filename[0:1] == "/" {
		fullPath = filepath.Clean(filename)
	} else if len(filename) > 0 && filename != "-a" {
		fullPath = filepath.Clean(conn.namePrefix + "/" + filename)
	} else {
		fullPath = filepath.Clean(conn.namePrefix)
	}
	fullPath = strings.Replace(fullPath, "//", "/", -1)
	fullPath = strings.Replace(fullPath, string(filepath.Separator), "/", -1)
	return
}

// sendOutofbandData will send a string to the client via the currently open
// data parallelSockets. Assumes the parallelSockets is open and ready to be used.
func (conn *Conn) sendOutofbandData(data []byte) {
	bytes := len(data)
	if conn.socket != nil {
		conn.socket.Write(data)
		conn.socket.Close()
		conn.socket = nil
	}
	message := "Closing data connection, sent " + strconv.Itoa(bytes) + " bytes"
	conn.writeMessage(226, message)
}

func (conn *Conn) sendOutofBandDataWriter(data io.ReadCloser) error {
	conn.lastFilePos = 0

	err := conn.sendDataOverSocket(data, conn.getActiveSocket())
	if err != nil {
		return err
	}

	conn.closeActiveSocket()

	return nil
}

func (conn *Conn) sendDataOverSocket(data io.Reader, socket socket.DataSocket) error {

	bytes, err := io.Copy(socket, data)

	if err != nil {
		return err
	}

	message := "Successfully sent " + strconv.Itoa(int(bytes)) + " bytes"
	conn.logger.Print(conn.sessionID, message)

	return nil
}

func (conn *Conn) sendDataOverSocketN(data io.Reader, socket socket.DataSocket, length int) error {

	bytes, err := io.CopyN(socket, data, int64(length))
	if err != nil {
		return err
	}

	message := "Successfully sent " + strconv.Itoa(int(bytes)) + " bytes"
	conn.writeMessage(200, message)

	return nil
}

func (conn *Conn) getActiveSocket() socket.DataSocket {

	if conn.extendedMode {
		return conn.parallelSockets
	} else {
		return conn.socket
	}

}

func (conn *Conn) closeActiveSocket() {

	message := "Closing data connection"
	conn.writeMessage(226, message)

	if conn.extendedMode {
		conn.parallelSockets.Close()
		conn.parallelSockets = nil
	} else {
		conn.socket.Close()
		conn.socket = nil
	}

}
