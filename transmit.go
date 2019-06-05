package scion

import (
	"encoding/gob"
	"github.com/jessevdk/go-flags"
	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/examples/transmit/client"
	"github.com/scionproto/scion/go/examples/transmit/scion"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var options struct {
	Local  string `short:"l" long:"local" required:"true" description:"Local address"`
	Remote string `short:"r" long:"remote" description:"Remote address"`
	Mode   string `short:"m" long:"mode" required:"true" choice:"server" choice:"client"`
}

type message struct {
	Data string
}

type quickStream struct {
	enc    *gob.Encoder
	dec    *gob.Decoder
	stream quic.Stream
}

func newQuickStream(stream quic.Stream) *quickStream {
	return &quickStream{
		enc:    gob.NewEncoder(stream),
		dec:    gob.NewDecoder(stream),
		stream: stream,
	}
}

func main() {

	parseFlags()

	if options.Mode == "server" {
		log.Debug("Running in server Mode")

		addr := parseAddress(options.Local)
		socket := scion.Listen(addr)

		for {

			session := socket.Accept()

			log.Debug("QUIC session accepted", "src", session.RemoteAddr())
			handleSession(session)
		}

	} else {

		log.Debug("Running in client Mode")

		local := parseAddress(options.Local)
		remote := parseAddress(options.Remote)

		conn := client.Connect(local, remote)

		stream := newQuickStream(conn.OpenStream())

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			for i := 0; i < 5; i++ {
				stream.sendMessage(message{"Hello World"})
				time.Sleep(500 * time.Millisecond)
			}
			wg.Done()
		}()

		go func() {

			for i := 0; i < 5; i++ {

				var msg message

				err := stream.dec.Decode(&msg)
				if err != nil {
					LogFatal("Unable to read", "err", err)
				}

				log.Debug("Received message", "msg", msg.Data)
			}

			wg.Done()
		}()

		wg.Wait()

		stream.close()
		conn.Close()

	}

}

func parseFlags() {
	_, err := flags.Parse(&options)

	if err != nil {
		LogFatal("Failed to parse flags", "err", err)
	}

}

func parseAddress(addr string) snet.Addr {

	local, err := snet.AddrFromString(addr)
	if err != nil {
		LogFatal("Failed to parse Local address", "err", err)
	}

	return *local
}

func handleSession(session quic.Session) {

	log.Debug("Now in handle session")

	stream, err := session.AcceptStream()

	if err != nil {
		LogFatal("Unable to accept stream", "err", err)
	}

	log.Debug("Accepted Stream", "id", stream.StreamID())

	qstream := newQuickStream(stream)

	decodeStream(qstream, session)
}

func decodeStream(stream *quickStream, session quic.Session) {

	for {

		var msg message
		err := stream.dec.Decode(&msg)

		if err != nil {

			// Reached EOF: Close Session
			if err.Error() == "EOF" {
				log.Debug("Received EOF, session terminated")
				session.Close(nil)
				stream.close()
				return
			}

			log.Debug(msg.Data)

			LogFatal("Unable to read", "err", err)
		}

		log.Debug("Received message", "msg", msg.Data)

		stream.sendMessage(message{"Bye World"})

	}
}

func (stream quickStream) sendMessage(msg message) {

	err := stream.enc.Encode(msg)
	if err != nil {
		LogFatal("Unable to send message", "err", err)
	}

	log.Debug("Sent message", "msg", msg.Data)
}

func (stream quickStream) receiveMessage() message {
	var msg message
	err := stream.dec.Decode(&msg)
	if err != nil {
		LogFatal("Unable to receive message", "err", err)
	}

	log.Debug("Received message", "msg", msg.Data)

	return msg
}

func (stream quickStream) close() {
	stream.stream.Close()
}

func LogFatal(msg string, a ...interface{}) {
	log.Crit(msg, a...)
	os.Exit(1)
}

func setSignalHandler(closer io.Closer) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		closer.Close()
		os.Exit(1)
	}()
}