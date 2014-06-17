// GoBalloon
// test-decode.go - APRS packet receiver + decoder for testing purposes
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"log"
	"net"
	"os"
	"os/signal"
	// "github.com/chrissnell/go-base91"
)

func main() {

	remote := flag.String("remote", "10.50.0.25:6701", "Remote TNC server")
	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	conn, err := net.Dial("tcp", *remote)
	if err != nil {
		log.Fatalf("Could not connect to %v.  Error: %v", *remote, err)
	}

	defer conn.Close()

	go func() {
		<-sig
		os.Exit(1)
	}()

	d := ax25.NewDecoder(conn)

	for {
		msg, err := d.Next()

		if err != nil {
			log.Printf("Error retrieving APRS message via KISS: %v", err)
		}

		fmt.Printf("%+v\n", msg)

	}
}
