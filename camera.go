// GoBalloon
// camera.go - Camera controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	_ "github.com/chrissnell/gpio"
	"time"
)

func CameraRun() {
	fmt.Println("--- runCamera start")
	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:
			//fmt.Println("--- Taking a photo")
			// pin := gpio.NewDigitalPin(12, "w")
			// pin.DigitalWrite("1")
			timer := time.NewTimer(time.Second * 1)
			<-timer.C
			// pin.DigitalWrite("0")
		}
	}
}
