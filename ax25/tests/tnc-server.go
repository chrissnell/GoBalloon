// GoBalloon
// tnc-server.go - A serial/TCP bridge for connecting to an AX.25 TNC device
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"github.com/tarm/goserial"
	"io"
	"log"
	"net"
)

func main() {

	port := flag.String("port", "/dev/ttyUSB0", "Serial port device (defaults to /dev/ttyUSB0)")
	flag.Parse()

	// Spin off a goroutine to watch for a SIGINT and die if we get one
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		os.Exit(1)
	}()

	sc := &serial.Config{Name: *port, Baud: 4800}

	s, err := serial.OpenPort(sc)
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", ":6700")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		log.Printf("Answered incoming connection from %v\n", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(s, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}

	err = s.Close()
	if err != nil {
		log.Fatalf("Error closing port: %v", err)
	}
}
