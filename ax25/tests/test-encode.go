package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/tarm/goserial"
	"log"
)

func main() {

	port := flag.String("port", "/dev/ttyUSB0", "Serial port device (defaults to /dev/ttyUSB0)")
	flag.Parse()

	c := &serial.Config{Name: *port, Baud: 4800}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

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

	s.Write(packet)

	err = s.Close()
	if err != nil {
		log.Fatalf("Error closing port: %v", err)
	}

	// Let's decode our own packet to make sure tht it's bueno
	buf := bytes.NewReader(packet)
	d := ax25.NewDecoder(buf)
	msg, err := d.Next()
	fmt.Printf("%+v\n", msg)

}
