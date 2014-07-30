// GoBalloon
// goballoon.go - Main controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/tv42/topic"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"syscall"
)

var (
	shutdownFlight   = make(chan bool)
	shutdownComplete = make(chan bool)
	aprsMessage      = make(chan string)
	aprsPosition     = make(chan geospatial.Point)
	currentPosition  geospatial.Point
	remotegps        *string
	remotetnc        *string
	localtncport     *string
	ballooncall      *string
	balloonssid      *string
	chasercall       *string
	chaserssid       *string
	beaconint        *string
	debug            *bool
	balloonAddr      ax25.APRSAddress
)

const (
	symbolTable rune = '/'
	symbolCode  rune = 'O'
)

func main() {

	remotegps = flag.String("remotegps", "10.50.0.21:2947", "Remote gpsd server")
	remotetnc = flag.String("remotetnc", "10.50.0.25:6700", "Remote TNC server")
	localtncport = flag.String("localtncport", "", "Local serial port for TNC, e.g. /dev/ttyUSB0")
	ballooncall = flag.String("ballooncall", "", "Balloon Callsign")
	balloonssid = flag.String("balloonssid", "", "Balloon SSID")
	chasercall = flag.String("chasercall", "", "Chaser Callsign")
	chaserssid = flag.String("chaserssid", "", "Chaser SSID")
	beaconint = flag.String("beaconint", "60", "APRS position beacon interval (secs)  Default: 60")
	debug = flag.Bool("debug", false, "Enable debugging information")

	flag.Parse()

	log.Println("Starting up.")

	f, _ := os.Create("goballoon.pprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6464", nil))
	}()

	if (len(*remotetnc) == 0) && (len(*localtncport) == 0) {
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

	// We're going to use Topic to handle the one -> many distribution
	// of our GPS readings
	top := topic.New()
	defer close(top.Broadcast)

	go CameraRun()
	go StartAPRSTNCConnector()
	go StartAPRSPositionBeacon(top)
	go GPSRun(top)
	<-sc
	shutdownFlight <- true
	close(shutdownFlight)
	log.Println("Shutting down.")

	// This needs to be converted to use sync.WaitGroup
	// <-shutdownComplete
	log.Println("Shutdown complete.")
}
