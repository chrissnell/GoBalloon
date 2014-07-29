// GoBalloon
// camera.go - Camera servo controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
)

func CameraRun() {
	fmt.Println("CameraRun() start")
	for {
		select {
		case <-shutdownFlight:
			fmt.Println("CameraRun() Break")
			return

			// NOT YET IMPLEMENTED

			// This block will activate the camera servo periodically, to point the camera
			// at the ground and the horizon.

		}
	}
}
