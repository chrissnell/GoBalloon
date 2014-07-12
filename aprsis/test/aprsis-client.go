// GoBalloon
// aprsis-client.go - An APRS-IS client.  Connects to APRS-IS and parses shit.

// (c) 2014, Christopher Snell

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/aprsis"
	"github.com/chrissnell/GoBalloon/ax25"
	"log"
	"net"
	"os"
)

var call, pass, filter, server, rawlog string

func init() {
	flag.StringVar(&server, "server", "second.aprs.net:14580", "APRS-IS upstream")
	flag.StringVar(&call, "call", "", "Your callsign (for APRS-IS)")
	flag.StringVar(&pass, "pass", "", "Your call pass (for APRS-IS)")
	flag.StringVar(&filter, "filter", "", "Optional filter for APRS-IS server")
	flag.StringVar(&rawlog, "rawlog", "", "Path to log raw messages")

}

func main() {
	flag.Parse()

	is, err := aprsis.Dial("tcp", server)
	if err != nil {
		log.Fatalln(err)
	}

	is.Auth(call, pass, filter)

	if rawlog != "" {
		logWriter, err := os.OpenFile(rawlog,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		is.SetRawLog(logWriter)
	}

	for {

		fmt.Println("-------------------------------------------")
		msg, err := is.Next()

		ad := aprs.ParsePacket(&msg)

		fmt.Println(msg)

		// If we have a recipient and a sender, add the sender to the APRSData.Message struct
		if ad.Message.Recipient.Callsign != "" && msg.Source.Callsign != "" {
			ad.Message.Sender = msg.Source
		}

		if ad.Position.Lat != 0 {
			fmt.Printf("Decoded APRS Data: %+v\n", ad)
		}

		if ad.Message.Recipient.String() != "" {
			fmt.Printf("%+v\n", msg)
		}

		if ad.Message.Recipient.String() == "NW5W-1" {
			fmt.Printf("Incoming message for NW5W-1: %+v\n", ad)
			fmt.Printf("%+v\n", msg)
			if ad.Message.ID != "" {
				fmt.Printf("ACKing message [%v]\n", ad.Message.ID)
				ackMessage(ad.Message)
			}
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Println("-------------------------------------------")

	}

}

func ackMessage(m aprs.Message) {

	psource := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     1,
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

	em, _ := aprs.CreateMessageACK(m)

	a := ax25.APRSPacket{
		Source: psource,
		Dest:   pdest,
		Path:   path,
		Body:   em,
	}

	packet, err := ax25.EncodeAX25Command(a)
	if err != nil {
		log.Fatalf("Unable to create packet: %v", err)
	}

	conn, err := net.Dial("tcp", "10.50.0.25:6700")
	if err != nil {
		log.Fatalf("Could not connect to %v.  Error: %v", "10.50.0.25:6700", err)
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
