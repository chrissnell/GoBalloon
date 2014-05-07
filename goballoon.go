// GoBalloon
// goballoon.go - Main controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	timeToDie = make(chan bool, 1)
)

func main() {
	fmt.Println("Starting up.")
	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	go CameraRun()
	go aprsBeacon()
	go GPSRun()
	<-sc
	timeToDie <- true
	fmt.Println("Shutting down.")
}
