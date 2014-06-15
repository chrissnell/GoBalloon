// GoBalloon
// aprs.go - APRS controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"time"
)

var (
	aprsMessage = make(chan string, 10)
)

func aprsBeacon() {
	fmt.Println("--- sendAPRSBeacon start")
	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:
			go processIncomingAPRSMessage()
			fmt.Println("--- Sending an APRS position beacon")
			if currentPosition.Lat != 0 || currentPosition.Lon != 0 {
				var amsg = fmt.Sprintf("hi lat=%v lon=%v alt=%v", currentPosition.Lat, currentPosition.Lon, currentPosition.Alt)
				aprsMessage <- amsg
			}
			timer := time.NewTimer(time.Second * 10)
			<-timer.C

		}
	}
}

func processIncomingAPRSMessage() {
	select {
	case newMessage := <-aprsMessage:
		fmt.Println("--- APRS Message: ", newMessage)
	}
}
