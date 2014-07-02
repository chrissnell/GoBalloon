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
	"github.com/chrissnell/GoBalloon/geospatial"
	"log"
	"net"
	_ "time"
)

func main() {

	remote := flag.String("remote", "10.50.0.25:6700", "Remote TNC server")
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

	c := aprs.CompressedTelemetryReport{
		A1:       7714,
		A2:       13,
		A3:       2,
		A4:       3006,
		A5:       429,
		Digital:  51,
		Sequence: 12,
		//Sequence: uint16(time.Now().Second()),
	}

	_, err := aprs.CreateCompressedTelemetryReport(&c)
	if err != nil {
		log.Fatalln("Could not create compressed telemetry report: ", err)
	}

	// path := []ax25.APRSAddress{
	// 	{
	// 		Callsign: "K9JEB",
	// 		SSID:     2,
	// 	},
	// }

	point := geospatial.NewPoint()
	point.Lat = 47.2111
	point.Lon = -122.4898
	point.Alt = 207

	position := aprs.CreateCompressedPosition(point, '/', 'O')
	//body := fmt.Sprint(position, "GoBalloon-Test", ctr)
	body := position + "GoBalloon-NotFlying"

	a := ax25.APRSData{
		Source: psource,
		Dest:   pdest,
		Path:   path,
		Body:   body,
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

	fmt.Printf("Packet -> %v\n", packet)

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
