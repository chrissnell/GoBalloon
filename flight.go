// GoBalloon
// flight.go - Flight controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"github.com/mrmorphic/hwio"
	"log"
	"time"
)

func InitiateCutdown() {
	var pin string = "P9.17"

	outputPin, err := hwio.GetPinWithMode(pin, hwio.OUTPUT)
	if err != nil {
		log.Printf("Error getting GPIO pin: %v\n", err)
	}

	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:
			log.Printf("--- %v high\n", pin)
			hwio.DigitalWrite(outputPin, hwio.HIGH)
			timer := time.NewTimer(time.Second * 1)
			<-timer.C
			log.Printf("--- %v low\n", pin)
			hwio.DigitalWrite(outputPin, hwio.LOW)
			timer = time.NewTimer(time.Second * 1)
			<-timer.C
		}
	}
}
