package main

import (
	"github.com/elwin/transmit/client"
	"github.com/scionproto/scion/go/lib/log"
	"time"
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

	err = conn.Login("admin", "123456")
	if err != nil {
		log.Error("Failed to authenticate", "err", err)
	}

	err = conn.Mode(ftp.ModeExtendedBlockMode)
	if err != nil {
		log.Error("Could not switch mode", "err", err)
	}

	err = conn.RetrMultipleConns()
	if err != nil {
		log.Error("Something failed", "err", err)
	}

	/*
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
	*/

	/*

		f2, err := os.Create("/home/elwin/ftp/result2.txt")
		if err != nil {
			log.Error("Creating file", "err", err)
		}

		reader := make([]*ftp.Response, 5)
		var wg sync.WaitGroup
		wg.Add(5)
		for i := 0; i < 5; i++ {

			offset := i * 30

			response, err := conn.Eret("yolo.txt", offset, 20)

			if err != nil {
				log.Error("Something went wrong!", "err", err)
			}

			reader[i] = response
			io.Copy(f2, response)
			response.Close()

			wg.Done()
		}

		wg.Wait()

	*/

	/*
		response, err := conn.Eret("yolo.txt", 5, 10)
		if err != nil {
			log.Error("failed to eret", "err", err)
		}
	*/

	/*
		for i := 0; i < 5; i++ {
			_, err = io.Copy(f2, reader[i])
			reader[i].Close()
			if err != nil {
				log.Error("Copy data", "err", err)
			}

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
