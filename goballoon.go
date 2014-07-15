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
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	timeToDie       = make(chan bool, 1)
	currentPosition geospatial.Point
	remotegps       *string
	remotetnc       *string
	localtncport    *string
	mycall          *string
	myssid          *string
	debug           *bool
	balloonAddr     ax25.APRSAddress
)

func main() {

	remotegps = flag.String("remotegps", "10.50.0.21:2947", "Remote gpsd server")
	remotetnc = flag.String("remotetnc", "10.50.0.25:6700", "Remote TNC server")
	localtncport = flag.String("localtncport", "", "Local serial port for TNC, e.g. /dev/ttyUSB0")
	mycall = flag.String("mycall", "", "Balloon Callsign")
	myssid = flag.String("myssid", "", "Balloon SSID")
	debug = flag.Bool("debug", false, "Enable debugging information")

	flag.Parse()

	fmt.Println("Starting up.")

	if (len(*remotetnc) == 0) && (len(*localtncport) == 0) {
		log.Fatalln("Must specify a local or remote TNC.  Use -h for help.")
	}

	if len(*mycall) == 0 {
		log.Fatalln("Must provide a balloon callsign.  Use -h for help.")
	}

	balloonAddr.Callsign = *mycall
	ssidInt, _ := strconv.Atoi(*myssid)
	balloonAddr.SSID = uint8(ssidInt)

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)

	go CameraRun()
	go StartAPRS()
	go GPSRun()
	go InitiateCutdown()
	<-sc
	timeToDie <- true
	fmt.Println("Shutting down.")
}
