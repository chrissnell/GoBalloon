// GoBalloon
// goballoon.go - Main controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/chrissnell/GoBalloon/gps"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var (
	shutdownFlight = make(chan bool)
	aprsMessage    = make(chan string)
	aprsPosition   = make(chan geospatial.Point)
	remotegps      *string
	remotetnc      *string
	localtncport   *string
	ballooncall    *string
	balloonssid    *string
	chasercall     *string
	chaserssid     *string
	beaconint      *string
	debug          *bool
	balloonAddr    ax25.APRSAddress
)

const (
	symbolTable rune = '/'
	symbolCode  rune = 'O'
)

func main() {

	// Set up a new GPS
	g := new(gps.GPS)

	// Set up a new TNC with our APRS symbol
	a := new(APRSTNC)
	a.symbolTable = symbolTable
	a.symbolCode = symbolCode

	var wg sync.WaitGroup

	g.Remotegps = flag.String("remotegps", "10.50.0.21:2947", "Remote gpsd server")
	a.Remotetnc = flag.String("remotetnc", "10.50.0.25:6700", "Remote TNC server")
	a.Localtncport = flag.String("localtncport", "", "Local serial port for TNC, e.g. /dev/ttyUSB0")
	ballooncall = flag.String("ballooncall", "", "Balloon Callsign")
	balloonssid = flag.String("balloonssid", "", "Balloon SSID")
	chasercall = flag.String("chasercall", "", "Chaser Callsign")
	chaserssid = flag.String("chaserssid", "", "Chaser SSID")
	a.Beaconint = flag.String("beaconint", "60", "APRS position beacon interval (secs)  Default: 60")
	debug = flag.Bool("debug", false, "Enable debugging information")

	flag.Parse()

	g.Debug = debug

	log.Println("Starting up.")

	if (len(*a.Remotetnc) == 0) && (len(*a.Localtncport) == 0) {
		log.Fatalln("Must specify a local or remote TNC.  Use -h for help.")
	}

	if len(*ballooncall) == 0 {
		log.Fatalln("Must provide a balloon callsign.  Use -h for help.")
	}

	if len(*chasercall) == 0 {
		log.Fatalln("Must provide a chaser callsign.  Use -h for help.")
	}

	balloonAddr.Callsign = *ballooncall
	ssidInt, _ := strconv.Atoi(*balloonssid)
	balloonAddr.SSID = uint8(ssidInt)

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)

	go FlightComputer(&g.Reading, &wg)
	go CameraRun()
	go g.StartGPS()
	a.gps = &g.Reading
	go a.StartAPRS()
	<-sc
	shutdownFlight <- true
	close(shutdownFlight)
	log.Println("Shutting down.")
	wg.Wait()
	log.Println("Shutdown complete.")
}
