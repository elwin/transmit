package ftp

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/elwin/transmit/scion"
	"github.com/scionproto/scion/go/lib/snet"

	"io"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// Login authenticates the client with specified user and password.
//
// "anonymous"/"anonymous" is a common user/password scheme for FTP servers
// that allows anonymous read-only accounts.
func (server *ServerConn) Login(user, password string) error {
	code, message, err := server.cmd(-1, "USER %s", user)
	if err != nil {
		return err
	}

	switch code {
	case StatusLoggedIn:
	case StatusUserOK:
		_, _, err = server.cmd(StatusLoggedIn, "PASS %s", password)
		if err != nil {
			return err
		}
	default:
		return errors.New(message)
	}

	// Switch to binary mode
	if _, _, err = server.cmd(StatusCommandOK, "TYPE I"); err != nil {
		return err
	}

	// Switch to UTF-8
	err = server.setUTF8()

	// If using implicit TLS, make data connections also use TLS
	if server.options.tlsConfig != nil {
		server.cmd(StatusCommandOK, "PBSZ 0")
		server.cmd(StatusCommandOK, "PROT P")
	}

	return err
}

// feat issues a FEAT FTP command to list the additional commands supported by
// the remote FTP server.
// FEAT is described in RFC 2389
func (server *ServerConn) feat() error {
	code, message, err := server.cmd(-1, "FEAT")
	if err != nil {
		return err
	}

	if code != StatusSystem {
		// The server does not support the FEAT command. This is not an
		// error: we consider that there is no additional feature.
		return nil
	}

	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, " ") {
			continue
		}

		line = strings.TrimSpace(line)
		featureElements := strings.SplitN(line, " ", 2)

		command := featureElements[0]

		var commandDesc string
		if len(featureElements) == 2 {
			commandDesc = featureElements[1]
		}

		server.features[command] = commandDesc
	}

	return nil
}

// setUTF8 issues an "OPTS UTF8 ON" command.
func (server *ServerConn) setUTF8() error {
	if _, ok := server.features["UTF8"]; !ok {
		return nil
	}

	code, message, err := server.cmd(-1, "OPTS UTF8 ON")
	if err != nil {
		return err
	}

	// Workaround for FTP servers, that does not support this option.
	if code == StatusBadArguments {
		return nil
	}

	// The ftpd "filezilla-server" has FEAT support for UTF8, but always returns
	// "202 UTF8 mode is always enabled. No need to send this command." when
	// trying to use it. That's OK
	if code == StatusCommandNotImplemented {
		return nil
	}

	if code != StatusCommandOK {
		return errors.New(message)
	}

	return nil
}

// epsv issues an "EPSV" command to get a port number for a data connection.
func (server *ServerConn) epsv() (port int, err error) {

	_, line, err := server.cmd(StatusExtendedPassiveMode, "EPSV")

	if err != nil {
		return
	}

	fmt.Println(line)

	start := strings.Index(line, "|||")
	end := strings.LastIndex(line, "|")
	if start == -1 || end == -1 {
		err = errors.New("invalid EPSV response format")
		return
	}
	port, err = strconv.Atoi(line[start+3 : end])
	return
}

// pasv issues a "PASV" command to get a port number for a data connection.
func (server *ServerConn) pasv() (host string, port int, err error) {
	_, line, err := server.cmd(StatusPassiveMode, "PASV")
	if err != nil {
		return
	}

	// PASV response format : 227 Entering Passive Mode (h1,h2,h3,h4,p1,p2).
	start := strings.Index(line, "(")
	end := strings.LastIndex(line, ")")
	if start == -1 || end == -1 {
		err = errors.New("invalid PASV response format")
		return
	}

	// We have to split the response string
	pasvData := strings.Split(line[start+1:end], ",")

	if len(pasvData) < 6 {
		err = errors.New("invalid PASV response format")
		return
	}

	// Let's compute the port number
	portPart1, err1 := strconv.Atoi(pasvData[4])
	if err1 != nil {
		err = err1
		return
	}

	portPart2, err2 := strconv.Atoi(pasvData[5])
	if err2 != nil {
		err = err2
		return
	}

	// Recompose port
	port = portPart1*256 + portPart2

	// Make the IP address to connect to
	host = strings.Join(pasvData[0:4], ".")
	return
}

// getDataConn returns a host, port for a new data connection
// it uses the best available method to do so
func (server *ServerConn) getDataConn() (snet.Addr, error) {

	if !server.options.disableEPSV && !server.skipEPSV {
		if port, err := server.epsv(); err == nil {

			host := scion.AddrToString(server.remote)
			addr, err := snet.AddrFromString(host + ":" + strconv.Itoa(port))
			return *addr, err
		}

		// if there is an error, skip EPSV for the next attempts
		server.skipEPSV = true
	}

	return snet.Addr{}, nil

	// return server.pasv()
}

func (server *ServerConn) getDataConns() ([]snet.Addr, error) {

	return server.spas()

}

// openDataConn creates a new FTP data connection.
func (server *ServerConn) openDataConn() (scion.Conn, error) {
	addr, err := server.getDataConn()

	if err != nil {
		return nil, err
	}

	laddr, err := snet.AddrFromString(local)
	if err != nil {
		return nil, err
	}

	conn, err := scion.Dial(*laddr, addr)

	return conn, err
}

func (server *ServerConn) openDataConns() ([]scion.Conn, error) {

	addrs, err := server.getDataConns()
	if err != nil {
		return nil, err
	}

	laddr, err := snet.AddrFromString(local)
	if err != nil {
		return nil, err
	}

	var conns []scion.Conn

	for _, addr := range addrs {
		conn, err := scion.Dial(*laddr, addr)
		if err != nil {
			return nil, err
		}

		conns = append(conns, conn)
	}

	return conns, nil
}

// cmd is a helper function to execute a command and check for the expected FTP
// return code
func (server *ServerConn) cmd(expected int, format string, args ...interface{}) (int, string, error) {
	err := server.dispatchCmd(format, args...)
	if err != nil {
		return 0, "", err
	}

	code, message, err := server.conn.ReadResponse(expected)
	if err == nil {
		server.logger.PrintResponse(code, message)
	}

	return code, message, err
}

func (server *ServerConn) dispatchCmd(format string, args ...interface{}) error {
	server.logger.PrintCommand(format, args...)
	_, err := server.conn.Cmd(format, args...)
	return err
}

// cmdDataConnFrom executes a command which require a FTP data connection.
// Issues a REST FTP command to specify the number of bytes to skip for the transfer.
func (server *ServerConn) cmdDataConnFrom(offset uint64, format string, args ...interface{}) (scion.Conn, error) {
	conn, err := server.openDataConn()
	if err != nil {
		return nil, err
	}

	if offset != 0 {
		_, _, err := server.cmd(StatusRequestFilePending, "REST %d", offset)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	err = server.dispatchCmd(format, args...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	code, msg, err := server.conn.ReadResponse(-1)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if code != StatusAlreadyOpen && code != StatusAboutToSend {
		conn.Close()
		return nil, &textproto.Error{Code: code, Msg: msg}
	}

	return conn, nil
}

// NameList issues an NLST FTP command.
func (server *ServerConn) NameList(path string) (entries []string, err error) {
	conn, err := server.cmdDataConnFrom(0, "NLST %s", path)
	if err != nil {
		return
	}

	r := &Response{conn: conn, c: server}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		entries = append(entries, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return entries, err
	}
	return
}

// List issues a LIST FTP command.
func (server *ServerConn) List(path string) (entries []*Entry, err error) {
	var cmd string
	var parser parseFunc

	if server.mlstSupported {
		cmd = "MLSD"
		parser = parseRFC3659ListLine
	} else {
		cmd = "LIST"
		parser = parseListLine
	}

	conn, err := server.cmdDataConnFrom(0, "%s %s", cmd, path)
	if err != nil {
		return
	}

	r := &Response{conn: conn, c: server}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	now := time.Now()
	for scanner.Scan() {
		entry, err := parser(scanner.Text(), now, server.options.location)
		if err == nil {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return
}

// ChangeDir issues a CWD FTP command, which changes the current directory to
// the specified path.
func (server *ServerConn) ChangeDir(path string) error {
	_, _, err := server.cmd(StatusRequestedFileActionOK, "CWD %s", path)
	return err
}

// ChangeDirToParent issues a CDUP FTP command, which changes the current
// directory to the parent directory.  This is similar to a call to ChangeDir
// with a path set to "..".
func (server *ServerConn) ChangeDirToParent() error {
	_, _, err := server.cmd(StatusRequestedFileActionOK, "CDUP")
	return err
}

// CurrentDir issues a PWD FTP command, which Returns the path of the current
// directory.
func (server *ServerConn) CurrentDir() (string, error) {
	_, msg, err := server.cmd(StatusPathCreated, "PWD")
	if err != nil {
		return "", err
	}

	start := strings.Index(msg, "\"")
	end := strings.LastIndex(msg, "\"")

	if start == -1 || end == -1 {
		return "", errors.New("unsuported PWD response format")
	}

	return msg[start+1 : end], nil
}

// FileSize issues a SIZE FTP command, which Returns the size of the file
func (server *ServerConn) FileSize(path string) (int64, error) {
	_, msg, err := server.cmd(StatusFile, "SIZE %s", path)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(msg, 10, 64)
}

// Retr issues a RETR FTP command to fetch the specified file from the remote
// FTP server.
//
// The returned ReadCloser must be closed to cleanup the FTP data connection.
func (server *ServerConn) Retr(path string) (*Response, error) {
	return server.RetrFrom(path, 0)
}

// RetrFrom issues a RETR FTP command to fetch the specified file from the remote
// FTP server, the server will not send the offset first bytes of the file.
//
// The returned ReadCloser must be closed to cleanup the FTP data connection.
func (server *ServerConn) RetrFrom(path string, offset uint64) (*Response, error) {
	conn, err := server.cmdDataConnFrom(offset, "RETR %s", path)
	if err != nil {
		return nil, err
	}

	return &Response{conn: conn, c: server}, nil
}

// Stor issues a STOR FTP command to store a file to the remote FTP server.
// Stor creates the specified file with the content of the io.Reader.
//
// Hint: io.Pipe() can be used if an io.Writer is required.
func (server *ServerConn) Stor(path string, r io.Reader) error {
	return server.StorFrom(path, r, 0)
}

// StorFrom issues a STOR FTP command to store a file to the remote FTP server.
// Stor creates the specified file with the content of the io.Reader, writing
// on the server will start at the given file offset.
//
// Hint: io.Pipe() can be used if an io.Writer is required.
func (server *ServerConn) StorFrom(path string, r io.Reader, offset uint64) error {
	conn, err := server.cmdDataConnFrom(offset, "STOR %s", path)
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, r)
	conn.Close()
	if err != nil {
		return err
	}

	_, _, err = server.conn.ReadResponse(StatusClosingDataConnection)
	return err
}

// Rename renames a file on the remote FTP server.
func (server *ServerConn) Rename(from, to string) error {
	_, _, err := server.cmd(StatusRequestFilePending, "RNFR %s", from)
	if err != nil {
		return err
	}

	_, _, err = server.cmd(StatusRequestedFileActionOK, "RNTO %s", to)
	return err
}

// Delete issues a DELE FTP command to delete the specified file from the
// remote FTP server.
func (server *ServerConn) Delete(path string) error {
	_, _, err := server.cmd(StatusRequestedFileActionOK, "DELE %s", path)
	return err
}

// RemoveDirRecur deletes a non-empty folder recursively using
// RemoveDir and Delete
func (server *ServerConn) RemoveDirRecur(path string) error {
	err := server.ChangeDir(path)
	if err != nil {
		return err
	}
	currentDir, err := server.CurrentDir()
	if err != nil {
		return err
	}

	entries, err := server.List(currentDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name != ".." && entry.Name != "." {
			if entry.Type == EntryTypeFolder {
				err = server.RemoveDirRecur(currentDir + "/" + entry.Name)
				if err != nil {
					return err
				}
			} else {
				err = server.Delete(entry.Name)
				if err != nil {
					return err
				}
			}
		}
	}
	err = server.ChangeDirToParent()
	if err != nil {
		return err
	}
	err = server.RemoveDir(currentDir)
	return err
}

// MakeDir issues a MKD FTP command to create the specified directory on the
// remote FTP server.
func (server *ServerConn) MakeDir(path string) error {
	_, _, err := server.cmd(StatusPathCreated, "MKD %s", path)
	return err
}

// RemoveDir issues a RMD FTP command to remove the specified directory from
// the remote FTP server.
func (server *ServerConn) RemoveDir(path string) error {
	_, _, err := server.cmd(StatusRequestedFileActionOK, "RMD %s", path)
	return err
}

// NoOp issues a NOOP FTP command.
// NOOP has no effects and is usually used to prevent the remote FTP server to
// close the otherwise idle connection.
func (server *ServerConn) NoOp() error {
	_, _, err := server.cmd(StatusCommandOK, "NOOP")
	return err
}

// Logout issues a REIN FTP command to logout the current user.
func (server *ServerConn) Logout() error {
	_, _, err := server.cmd(StatusReady, "REIN")
	return err
}

// Quit issues a QUIT FTP command to properly close the connection from the
// remote FTP server.
func (server *ServerConn) Quit() error {
	server.dispatchCmd("QUIT")

	// Otherwise data connection will be closed before data is even sent
	time.Sleep(100 * time.Millisecond)

	return server.conn.Close()
}

// Extensions

func (server *ServerConn) spas() ([]snet.Addr, error) {
	_, line, err := server.cmd(StatusExtendedPassiveMode, "SPAS")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(line, "\n")

	var addrs []snet.Addr

	for _, line = range lines {
		if !strings.HasPrefix(line, " ") {
			continue
		}

		addr, err := snet.AddrFromString(strings.TrimLeft(line, " "))
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, *addr)
	}

	return addrs, nil
}

func (server *ServerConn) Eret(path string, offset, length int) error {

	conns, err := server.openDataConns()

	if err != nil {
		for _, conn := range conns {
			conn.Close()
		}

		return nil
	}

	_, line, err := server.cmd(StatusExtendedPassiveMode, "ERET %s=\"%d,%d\" %s")
	if err != nil {
		for _, conn := range conns {
			conn.Close()
		}

		return nil
	}
	fmt.Println(line)

	return nil
}

// Read implements the io.Reader interface on a FTP data connection.
func (r *Response) Read(buf []byte) (int, error) {
	return r.conn.Read(buf)
}

// Close implements the io.Closer interface on a FTP data connection.
// After the first call, Close will do nothing and return nil.
func (r *Response) Close() error {
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
func (r *Response) SetDeadline(t time.Time) error {
	return r.conn.SetDeadline(t)
}
