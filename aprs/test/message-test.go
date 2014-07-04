// GoBalloon
// tnc-client.go - An APRS encoder and sender, intended to talk to
//                 tnc-server, serial/TCP bridge for connecting to an AX.25 TNC device.
//
// (c) 2014, Christopher Snell

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"log"
	"net"
	_ "time"
)

func main() {

	remote := flag.String("remote", "10.50.0.25:6700", "Remote TNC server")
	sendit := flag.Bool("sendit", false, "Send message to RF (default: false)")
	flag.Parse()

	psource := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     7,
	}

	pdest := ax25.APRSAddress{
		Callsign: "APZ001",
		SSID:     0,
	}

	path := []ax25.APRSAddress{
		{
			Callsign: "WIDE1",
			SSID:     1,
		},
		{
			Callsign: "WIDE2",
			SSID:     1,
		},
	}

	sender := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     7,
	}

	recipient := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     1,
	}

	m := aprs.Message{
		Sender:    sender,
		Recipient: recipient,
		ID:        "001",
		Text:      "Testing 1 2 3",
	}

	ms, err := aprs.CreateMessage(&m)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if *sendit {

		a := ax25.APRSPacket{
			Source: psource,
			Dest:   pdest,
			Path:   path,
			Body:   ms,
		}

		packet, err := ax25.EncodeAX25Command(a)
		if err != nil {
			log.Fatalf("Unable to create packet: %v", err)
		}

		conn, err := net.Dial("tcp", *remote)
		if err != nil {
			log.Fatalf("Could not connect to %v.  Error: %v", *remote, err)
		}

		bw, err := conn.Write(packet)
		if err != nil {
			log.Fatalf("Could not write to remote.  Error: %v", err)
		} else {
			log.Printf("Wrote %v bytes to %v", bw, conn.RemoteAddr())
		}

		err = conn.Close()
		if err != nil {
			log.Fatalf("Error closing connection: %v", err)
		}

		// Let's decode our own packet to make sure tht it's bueno
		buf := bytes.NewReader(packet)
		d := ax25.NewDecoder(buf)
		msg, err := d.Next()
		fmt.Printf("%+v\n", msg)

	}

	fmt.Printf("Encoded message: %v\n", ms)

	dm, remains, err := aprs.DecodeMessage(ms)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Decoded message: %+v\n", dm)
	fmt.Printf("Remains: %v\n", remains)

	dm, remains, err = aprs.DecodeMessage(":NW5W-7   :ACK707")

	fmt.Printf("Decoded message: %+v\n", dm)
	fmt.Printf("Remains: %v\n", remains)

}
