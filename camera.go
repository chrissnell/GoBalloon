// GoBalloon
// camera.go - Camera servo controller code
//
// (c) 2014, Christopher Snell

package main

import (
	_ "github.com/chrissnell/gpio"
	"log"
	"time"
)

func CameraRun() {
	fmt.Println("CameraRun() start")
	for {
		select {
		case <-shutdownFlight:
			fmt.Println("CameraRun() Break")
			break
		default:

			// NOT YET IMPLEMENTED

			// This block will activate the camera servo periodically, to point the camera
			// at the ground and the horizon.

		}
	}
}
