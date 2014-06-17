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
	"os"
	"os/signal"
)

func serialWriterNetReader(netconn net.Conn, serialconn io.ReadWriteCloser, serialWriterDone chan bool) {
	b, err := io.Copy(serialconn, netconn)
	if err != nil {
		log.Printf("Error copying from network to serial: %v", err)
		log.Printf("netToSerial connection closing.  %v bytes written.", b)
		return
	}
	netconn.Close()
	serialWriterDone <- true
	log.Printf("netToSerial connection closing.  %v bytes written.", b)
	return
}

func netWriterSerialReader(netconn net.Conn, serialconn io.ReadWriteCloser, serialReaderDone chan bool) {
	b, err := io.Copy(netconn, serialconn)
	if err != nil {
		log.Printf("Error copying from serial to network: %v", err)
		return
	}
	netconn.Close()
	serialReaderDone <- true
	log.Printf("serialToNet connection closing.  %v bytes written.", b)
	return
}

func waitForSerialWriter(netToSerialListener net.Listener, s io.ReadWriteCloser) {

	for {
		serialWriterDone := make(chan bool, 1)

		// Wait for a connection.
		conn, err := netToSerialListener.Accept()
		log.Printf("Answered incoming Writer connection from %v\n", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}

		go serialWriterNetReader(conn, s, serialWriterDone)

		<-serialWriterDone
	}

}

func waitForNetWriter(serialToNetListener net.Listener, s io.ReadWriteCloser) {

	for {
		serialReaderDone := make(chan bool, 1)

		// Wait for a connection.
		conn, err := serialToNetListener.Accept()
		log.Printf("Answered incoming Reader connection from %v\n", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}

		go netWriterSerialReader(conn, s, serialReaderDone)

		<-serialReaderDone
	}

}

func main() {

	port := flag.String("port", "/dev/ttyUSB0", "Serial port device (defaults to /dev/ttyUSB0)")
	flag.Parse()

	// Spin off a goroutine to watch for a SIGINT and die if we get one
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	//go func() {
	//	<-sig
	//	os.Exit(1)
	//}()

	sc := &serial.Config{Name: *port, Baud: 4800}

	s, err := serial.OpenPort(sc)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	netToSerialListener, err := net.Listen("tcp", ":6700")
	if err != nil {
		log.Fatal(err)
	}
	defer netToSerialListener.Close()

	serialToNetListener, err := net.Listen("tcp", ":6701")
	if err != nil {
		log.Fatal(err)
	}
	defer serialToNetListener.Close()

	go waitForSerialWriter(netToSerialListener, s)
	go waitForNetWriter(serialToNetListener, s)

	<-sig
	log.Println("SIGINT received.  Shutting down...")
	os.Exit(1)
}
