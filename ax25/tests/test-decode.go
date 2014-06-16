// GoBalloon
// test-decode.go - APRS packet receiver + decoder for testing purposes
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/tarm/goserial"
	"log"
	"os"
	"os/signal"
	// "github.com/chrissnell/go-base91"
)

func main() {

	port := flag.String("port", "/dev/ttyUSB0", "Serial port device (defaults to /dev/ttyUSB0)")
	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	c := &serial.Config{Name: *port, Baud: 4800}
	//b := make([]byte, 1)

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-sig
		os.Exit(1)
	}()

	d := ax25.NewDecoder(s)

	for {
		msg, err := d.Next()

		if err != nil {
			log.Printf("Error retrieving APRS message via KISS: %v", err)
		}

		fmt.Printf("%+v\n", msg)

	}
}
