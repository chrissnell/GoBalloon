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
	"log"
	"net"
	"time"
)

var (
	aprsMessage = make(chan string, 10)
)

func aprsBeacon(aprssource, aprsdest ax25.APRSAddress) {
	fmt.Println("--- sendAPRSBeacon start")

	var path []ax25.APRSAddress

	path = append(path, ax25.APRSAddress{
		Callsign: "WIDE1",
		SSID:     1,
	})

	path = append(path, ax25.APRSAddress{
		Callsign: "WIDE2",
		SSID:     1,
	})

	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:
			go processIncomingAPRSMessage()
			fmt.Println("--- Sending an APRS position beacon")
			if currentPosition.Lat != 0 || currentPosition.Lon != 0 {
				var amsg = fmt.Sprintf("hi lat=%v lon=%v alt=%v", currentPosition.Lat, currentPosition.Lon, currentPosition.Alt)
				aprsMessage <- amsg

				point := geospatial.NewPoint()
				point.Lat = currentPosition.Lat
				point.Lon = currentPosition.Lon
				point.Alt = currentPosition.Alt

				position := aprs.CreateCompressedPosition(point, '/', 'O')
				body := fmt.Sprint(position, "GoBalloon Test http://nw5w.com")

				a := ax25.APRSData{
					Source: aprssource,
					Dest:   aprsdest,
					Path:   path,
					Body:   body,
				}

				packet, err := ax25.EncodeAX25Command(a)
				if err != nil {
					log.Fatalf("Unable to create packet: %v", err)
				}

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

				// Let's decode our own packet to make sure tht it's bueno
				buf := bytes.NewReader(packet)
				d := ax25.NewDecoder(buf)
				msg, err := d.Next()
				fmt.Printf("%+v\n", msg)

			}
			timer := time.NewTimer(time.Second * 30)
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
