// GoBalloon
// flight.go - Flight controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	_ "github.com/chrissnell/gpio"
	"github.com/mrmorphic/hwio"
	"log"
	"time"
)

func InitiateCutdown() {
	// P8-10
	// GPIO_68 on poinout diagrams
	// (Power/reset side of board, 5th row down from power/reset, on outside column)

	// Valid pins:  GPIO2_3 (pin 8, P8)
	//				GPIO2_4 (pin 10, P8)
	//				GPIO2_2	(pin 7, P8)
	//				GPIO1_13 (pin 11, P8)
	var pin string = "gpio1_13"

	outputPin, err := hwio.GetPinWithMode(pin, hwio.OUTPUT)
	if err != nil {
		log.Printf("Error getting GPIO pin: %v\n", err)
	}

	aprsMessage <- "Preparing to cutdown in 10 sec"
	timer := time.NewTimer(time.Second * 10)
	<-timer.C
	fmt.Println("--- CUTTING DOWN ---")
	hwio.DigitalWrite(outputPin, hwio.HIGH)
	timer = time.NewTimer(time.Second * 10)
	<-timer.C
	hwio.DigitalWrite(outputPin, hwio.LOW)
	hwio.CloseAll()
	fmt.Println("Closed all pins")

}

func SoundBuzzer() {

}
