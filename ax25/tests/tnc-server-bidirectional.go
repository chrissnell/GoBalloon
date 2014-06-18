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

func serialWriter(netconn net.Conn, serialconn io.ReadWriteCloser, serialWriterDone chan bool) {
	b, err := io.Copy(serialconn, netconn)
	if err != nil {
		log.Printf("Error copying from network->serial: %v", err)
		log.Printf("serialWriter copy closing.  %v bytes written.", b)
		serialWriterDone <- true
		return
	}
	netconn.Close()
	serialWriterDone <- true
	log.Printf("serialWriter connection closing.  %v bytes written.", b)
	return
}

func serialReader(netconn net.Conn, serialconn io.ReadWriteCloser, serialReaderDone chan bool) {
	b, err := io.Copy(netconn, serialconn)
	if err != nil {
		log.Printf("Error copying from serial->network: %v", err)
		log.Printf("serialReader copy closing.  %v bytes written.", b)
		serialReaderDone <- true
		return
	}
	serialReaderDone <- true
	log.Printf("serialReader connection closing.  %v bytes written.", b)
	return
}

func waitForSerialWriter(serialWriterListener net.Listener, s io.ReadWriteCloser) {

	log.Println("Starting serialWriterListener...")

	for {
		serialWriterDone := make(chan bool, 1)

		// Wait for a connection.
		conn, err := serialWriterListener.Accept()
		log.Printf("Answered incoming Writer connection from %v\n", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}

		go serialWriter(conn, s, serialWriterDone)

		<-serialWriterDone
		log.Println("Serial Writer done.")
	}

}

func waitForSerialReader(serialReaderListener net.Listener, s io.ReadWriteCloser) {

	log.Println("Starting serialReaderListener...")

	for {
		serialReaderDone := make(chan bool, 1)

		// Wait for a connection.
		conn, err := serialReaderListener.Accept()
		log.Printf("Answered incoming Reader connection from %v\n", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}

		go serialReader(conn, s, serialReaderDone)

		<-serialReaderDone
		conn.Close()
		log.Println("Serial Reader done.")

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

	serialWriterListener, err := net.Listen("tcp", ":6700")
	if err != nil {
		log.Fatal(err)
	}
	defer serialWriterListener.Close()

	serialReaderListener, err := net.Listen("tcp", ":6701")
	if err != nil {
		log.Fatal(err)
	}
	defer serialReaderListener.Close()

	go waitForSerialWriter(serialWriterListener, s)
	go waitForSerialReader(serialReaderListener, s)

	<-sig
	log.Println("SIGINT received.  Shutting down...")
	os.Exit(1)
}
