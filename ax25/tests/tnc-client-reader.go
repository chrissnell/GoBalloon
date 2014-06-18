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
	_ "time"
	// "github.com/chrissnell/go-base91"
)

func main() {

	remote := flag.String("remote", "10.50.0.25:6701", "Remote TNC server")
	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		os.Exit(1)
	}()

	for {

		log.Printf("Connecting to %v...\n", *remote)
		conn, err := net.Dial("tcp", *remote)
		if err != nil {
			log.Fatalf("Could not connect to %v.  Error: %v", *remote, err)
		}

		d := ax25.NewDecoder(conn)

		for {
			msg, err := d.Next()

			if err != nil {
				log.Printf("Error retrieving APRS message via KISS: %v", err)
				// log.Println("Sleeping 5 seconds and reconnecting.")
				// timer := time.NewTimer(time.Second * 5)
				// <-timer.C
				// break

			}

			fmt.Printf("%+v\n", msg)
		}

		log.Println("Closing connection.")
		conn.Close()

	}
}
