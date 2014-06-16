// GoBalloon
// aprs.go - APRS controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"bytes"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/tarm/goserial"
	"log"
	"net"
	"time"
)

var (
	aprsMessage = make(chan string, 10)
)

func aprsBeacon(aprssource, aprsdest ax25.APRSAddress) {
	fmt.Println("--- sendAPRSBeacon start")

	var a ax25.APRSData

	lowpath := []ax25.APRSAddress{
		{
			Callsign: "WIDE1",
			SSID:     1,
		},
		{
			Callsign: "WIDE2",
			SSID:     1,
		},
	}

	highpath := []ax25.APRSAddress{{
		Callsign: "WIDE2",
		SSID:     1,
	}}

	// This loop runs until a message on thetimeToDie channel is received
	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:

			// Start our incoming message processor in a separate goroutine
			go processIncomingAPRSMessage()

			// Only transmit packets when the GPS has satellite lock
			if currentPosition.Lat != 0 || currentPosition.Lon != 0 {

				// This is how we handle incoming APRS messages.  Not yet implemented
				//var amsg = fmt.Sprintf("hi lat=%v lon=%v alt=%v", currentPosition.Lat, currentPosition.Lon, currentPosition.Alt)
				//aprsMessage <- amsg

				// Store our current position in a geospatial.Point
				point := geospatial.NewPoint()
				point.Lat = currentPosition.Lat
				point.Lon = currentPosition.Lon
				point.Alt = currentPosition.Alt

				// Create a compressed position report from that point, using the Balloon symbol
				position := aprs.CreateCompressedPosition(point, '/', 'O')

				// Append our comment to the position report
				body := fmt.Sprint(position, "GoBalloon Test http://nw5w.com")

				// We use a much smaller path when flying above 5000' MSL
				if currentPosition.Alt > 5000 {

					// Form an APRS data packet
					a = ax25.APRSData{
						Source: aprssource,
						Dest:   aprsdest,
						Path:   highpath,
						Body:   body,
					}
				} else {

					// Form an APRS data packet
					a = ax25.APRSData{
						Source: aprssource,
						Dest:   aprsdest,
						Path:   lowpath,
						Body:   body,
					}
				}

				packet, err := ax25.EncodeAX25Command(a)
				if err != nil {
					log.Fatalf("Unable to create packet: %v", err)
				}

				fmt.Println("--- Sending an APRS position beacon")

				if len(*remotetnc) == 0 {

					// We weren't passed a remote TNC so we'll connect to a local one.
					if len(*localtncport) == 0 {
						log.Fatalln("No remote TNC address or local TNC port were specified.")
					}

					sc := &serial.Config{Name: *localtncport, Baud: 4800}
					s, err := serial.OpenPort(sc)
					if err != nil {
						log.Fatal(err)
					}

					// Send our KISS packet over the serial port
					s.Write(packet)

					err = s.Close()
					if err != nil {
						log.Fatalf("Error closing port: %v", err)
					}

				} else {

					conn, err := net.Dial("tcp", *remotetnc)
					if err != nil {
						log.Fatalf("Could not connect to %v.  Error: %v", *remotetnc, err)
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
				}

				// Let's decode our own packet to make sure tht it's bueno
				buf := bytes.NewReader(packet)
				d := ax25.NewDecoder(buf)
				msg, err := d.Next()
				fmt.Printf("%+v\n", msg)

			}
			timer := time.NewTimer(time.Second * 300)
			<-timer.C

		}
	}
}

func processIncomingAPRSMessage() {
	select {
	case newMessage := <-aprsMessage:
		fmt.Println("--- APRS Message: ", newMessage)
	}
}
