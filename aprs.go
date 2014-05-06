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
			fmt.Println("--- Sending an APRS beacon")
			var amsg = "hi"
			aprsMessage <- amsg
			timer := time.NewTimer(time.Second * 10)
			<-timer.C

		}
	}
}

func processIncomingAPRSMessage() {
	select {
	case newMessage := <-aprsMessage:
		fmt.Println(newMessage)
	}
}
