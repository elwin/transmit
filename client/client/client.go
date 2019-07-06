package main

import (
	"fmt"
	"github.com/elwin/transmit/mode"
	"os"
	"time"

	"github.com/elwin/transmit/client"
	"github.com/scionproto/scion/go/lib/log"
)

func main() {

	conn, err := ftp.Dial(
		"1-ff00:0:110,[127.0.0.1]:2121",
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

	entries, _ := conn.List("/")
	for _, entry := range entries {
		fmt.Println(entry.Name)
	}

	err = conn.Mode(mode.ExtendedBlockMode)
	if err != nil {
		log.Error("Could not switch mode", "err", err)
	}

	f, _ := os.Open("/home/elwin/ftp/yolo.txt")

	err = conn.Stor("stor.txt", f)
	if err != nil {
		log.Error("Something happened when wiriting", "err", err)
	}

	/*
		response, err := conn.Retr("yolo.txt")
		if err != nil {
			log.Error("Something failed", "err", err)
		}

		f, _ := os.Create("/home/elwin/ftp/result.txt")
		_, err = io.Copy(f, response)
		// response.Close()


	*/

	entries, err = conn.List("/")
	if err != nil {
		log.Error("List", "err", err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name)
	}
}
