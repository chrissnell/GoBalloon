package main

import (
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/tarm/goserial"
	"log"
	//	"os"
	//	"os/signal"
	// "github.com/chrissnell/go-base91"
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

	ppath := ax25.APRSAddress{
		Callsign: "WIDE2",
		SSID:     2,
	}

	a := ax25.APRSData{
		Source: psource,
		Dest:   pdest,
		Path:   []ax25.APRSAddress{ppath},
		Body:   "!4715.68N/12228.20W-GoBalloon Test http://nw5w.com",
	}

	packet, err := ax25.EncodeAX25Command(a)
	if err != nil {
		log.Fatalf("Unable to create packet: %v", err)
	} else {
		fmt.Printf("--> %v\n", string(packet))
	}

	s.Write(packet)

	err = s.Close()
	if err != nil {
		log.Fatalf("Error closing port: %v", err)
	}

}
