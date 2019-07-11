package main

import (
	"flag"
	"fmt"
	"github.com/elwin/transmit/mode"
	l "log"
	"strings"
	"time"

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

	err = conn.Stor("stor1.txt", strings.NewReader("Hello World!\n"))
	if err != nil {
		log.Error("failed to stor", "err", err)
	}

	err = conn.Stor("stor2.txt", strings.NewReader("Bye World!\n"))
	if err != nil {
		log.Error("failed to stor", "err", err)
	}

	entries, _ := conn.List("/")
	for _, entry := range entries {
		fmt.Println(entry.Name)
	}

	conn.Mode(mode.ExtendedBlockMode)

	response, _ := conn.Retr("stor1.txt")
	buf := make([]byte, 10)

	for {
		n, _ := response.Read(buf)
		if n == 0 {
			break
		}

		fmt.Print(string(buf[0:n]))
	}

	/*

		entries, _ := conn.List("/")
		for _, entry := range entries {
			fmt.Println(entry.Name)
		}

		err = conn.Mode(mode.ExtendedBlockMode)
		if err != nil {
			log.Error("Could not switch mode", "err", err)
		}

		for i := 0; i < 2; i++ {

			name := "b" + strconv.Itoa(i) + ".txt"

			response, err := conn.Retr("a.txt")
			if err != nil {
				log.Error("Something failed", "err", err)
			}

			f, _ := os.Create("/home/elwin/ftp/" + name)
			_, err = io.Copy(f, response)
			response.Close()

			conn.ChangeDir("sub")

			f, _ = os.Open("/home/elwin/ftp/a.txt")

			err = conn.Stor(name, f)
			if err != nil {
				log.Error("Something happened when writing", "err", err)
			}

			conn.ChangeDirToParent()
		}
	*/

	// Send file back

	/*
		entries, err = conn.List("/")
		if err != nil {
			log.Error("List", "err", err)
		}
		for _, entry := range entries {
			fmt.Println(entry.Name)
		}

	*/

}
