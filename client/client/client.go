package main

import (
	"github.com/elwin/transmit/client"
	"github.com/scionproto/scion/go/lib/log"
	"io"
	"os"
	"strings"
	"time"
)

func main() {

	conn, err := ftp.Dial(
		"1-ff00:0:110,[127.0.0.1]:2121",
		// ftp.DialWithDebugOutput(os.Stdout),
		ftp.DialWithTimeout(60*time.Second),
	)

	if err != nil {
		log.Error("Failed to dial", "msg", err)
	}

	err = conn.Login("admin", "123456")
	if err != nil {
		log.Error("Failed to authenticate", "msg", err)
	}

	conn.Stor("yolo.txt", strings.NewReader("This data is supposed to be sent and retrieved subsequently"))

	response, err := conn.Retr("yolo.txt")
	if err != nil {
		log.Error("Retr", "err", err)
	}

	f, err := os.Create("/home/elwin/ftp/result.txt")
	if err != nil {
		log.Error("Creating file", "err", err)
	}

	_, err = io.Copy(f, response)

	if err != nil {
		log.Error("Copy data", "err", err)
	}

	/*
		err = conn.Spas()
		if err != nil {
			log.Error("Yo spas", "err", err)
		}
	*/

	conn.Quit()
}

/*

	c.Stor("test1.txt", strings.NewReader("My message"))
	c.Stor("test2.txt", strings.NewReader("Bye World"))

	c.MakeDir("yolodir")
	c.ChangeDir("yolodir")
	c.Stor("something.txt", strings.NewReader("This is some fancy new stuff"))
	c.ChangeDirToParent()

*/
