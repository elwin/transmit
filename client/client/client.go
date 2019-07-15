package main

import (
	"flag"
	"io"
	l "log"
	"strings"
	"time"

	"github.com/elwin/transmit/mode"

	"github.com/elwin/transmit/client"
	"github.com/scionproto/scion/go/lib/log"
)

func main() {

	var (
		local  = flag.String("local", "", "Local address (Format: AS,[IP])")
		remote = flag.String("remote", "", "Remote address to connect to (Format: AS,[IP]:Port)")
	)

	flag.Parse()
	if *local == "" {
		l.Fatalf("Please set a local address with -local")
	}
	if *remote == "" {
		l.Fatalf("Please set a remote address with -remote")
	}

	conn, err := ftp.Dial(
		*local,
		*remote,
		// ftp.DialWithDebugOutput(os.Stdout),
		ftp.DialWithTimeout(60*time.Second),
	)
	if err != nil {
		log.Error("Failed to dial", "err", err)
	}
	defer conn.Quit()

	err = conn.Login("admin", "123456")
	if err != nil {
		log.Error("Failed to authenticate", "err", err)
	}

	conn.Mode(mode.ExtendedBlockMode)

	reader := &myReader{strings.NewReader("Hello World")}

	err = conn.Stor("stor1.txt", reader)
	if err != nil {
		log.Error("failed to stor", "err", err)
	}

	/*

		err = conn.Stor("stor2.txt", strings.NewReader("Bye World!\n"))
		if err != nil {
			log.Error("failed to stor", "err", err)
		}

		entries, _ := conn.List("/")
		for _, entry := range entries {
			fmt.Println(entry.Name)
		}

		response, _ := conn.Retr("retr.txt")
		os.Mkdir("ftp", os.ModePerm)
		f, err := os.Create("ftp/retr.txt")
		if err != nil {
			log.Error("failed to create file", "err", err)
		}
		io.Copy(f, response)

		response.Close()

	*/

}

// Prevent io.Copy from using
var _ io.Reader = &myReader{}

type myReader struct {
	r io.Reader
}

func (w *myReader) Read(p []byte) (n int, err error) {
	return w.r.Read(p)
}
