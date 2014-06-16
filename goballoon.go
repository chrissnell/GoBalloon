// GoBalloon
// goballoon.go - Main controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"os"
	"os/signal"
	"syscall"
)

var (
	timeToDie       = make(chan bool, 1)
	currentPosition geospatial.Point
	remotegps       *string
	remotetnc       *string
)

func main() {

	remotegps = flag.String("remotegps", "10.50.0.21:2947", "Remote gpsd server")
	remotetnc = flag.String("remote", "10.50.0.25:6700", "Remote TNC server")

	flag.Parse()

	fmt.Println("Starting up.")

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)

	aprssource := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     7,
	}

	aprsdest := ax25.APRSAddress{
		Callsign: "APZ001",
		SSID:     0,
	}

	go CameraRun()
	go aprsBeacon(aprssource, aprsdest)
	go GPSRun()
	<-sc
	timeToDie <- true
	fmt.Println("Shutting down.")
}
