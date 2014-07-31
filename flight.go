// GoBalloon
// flight.go - Flight controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"github.com/mrmorphic/hwio"
	"log"
	"sync"
	"time"
)

func FlightComputer(g *GPSReading, wg *sync.WaitGroup) {

	var maxalt float64
	var once sync.Once
	var timer *time.Timer

	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-shutdownFlight:
			return

		default:
			pos := g.Get()
			if pos.Lat != 0 && pos.Lon != 0 {
				if pos.Altitude > maxalt {
					maxalt = pos.Altitude
				}
				log.Printf("MAX ALT: %v\n", maxalt)

				if maxalt > 17000 && pos.Altitude < 15000 {
					once.Do(func() { SoundBuzzer(wg) })
				}

			}

			timer = time.NewTimer(time.Second * 5)
			<-timer.C

		}

	}

}

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
		log.Printf("InitiateCutdown() :: Error getting GPIO pin: %v\n", err)
	}

	aprsMessage <- "Preparing to cutdown in 30 sec"
	timer := time.NewTimer(time.Second * 30)
	<-timer.C
	log.Println("--- CUTTING DOWN ---")
	hwio.DigitalWrite(outputPin, hwio.HIGH)
	timer = time.NewTimer(time.Second * 10)
	<-timer.C
	hwio.DigitalWrite(outputPin, hwio.LOW)
	hwio.CloseAll()
	log.Println("InitiateCutdown() :: Closed all pins")

}

func SoundBuzzer(wg *sync.WaitGroup) {

	var timer, timer2 *time.Timer
	var pin string = "gpio2_2"
	toggle := make(chan bool)

	wg.Add(1)
	defer wg.Done()

	log.Println("Activating buzzer")

	outputPin, err := hwio.GetPinWithMode(pin, hwio.OUTPUT)
	if err != nil {
		log.Printf("Error getting GPIO pin: %v\n", err)
	}

	go func() {
		for {
			timer = time.NewTimer(time.Millisecond * 1000)
			<-timer.C
			toggle <- true
			timer2 = time.NewTimer(time.Millisecond * 50)
			<-timer2.C
			toggle <- false
		}
	}()

	for {
		select {
		case <-shutdownFlight:
			log.Println("SoundBuzzer() :: Break")
			hwio.DigitalWrite(outputPin, hwio.LOW)
			hwio.CloseAll()
			log.Println("SoundBuzzer() :: Closed all pins")
			return
		case t := <-toggle:
			log.Printf("SoundBuzzer() :: Toggling buzzer: %v\n", t)
			if t {
				hwio.DigitalWrite(outputPin, hwio.HIGH)
			} else {
				hwio.DigitalWrite(outputPin, hwio.LOW)
			}
		}

	}

}
