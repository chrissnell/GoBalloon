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

	var path []ax25.APRSAddress

	path = append(path, ax25.APRSAddress{
		Callsign: "WIDE1",
		SSID:     1,
	})

	path = append(path, ax25.APRSAddress{
		Callsign: "WIDE2",
		SSID:     1,
	})

	point := geospatial.NewPoint()
	point.Lat = 47.262347
	point.Lon = -122.46988
	point.Alt = 1702

	position := aprs.CreateCompressedPosition(point, '/', 'O')
	body := fmt.Sprint(position, "GoBalloon Test http://nw5w.com")

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
