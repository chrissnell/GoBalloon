package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	timeToDie   = make(chan bool, 1)
	aprsMessage = make(chan string, 10)
)

func main() {
	fmt.Println("Starting up.")
	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	go runCamera()
	go aprsBeacon()
	<-sc
	timeToDie <- true
	fmt.Println("Shutting down.")
}

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

func runCamera() {
	fmt.Println("--- runCamera start")
	for {
		select {
		case <-timeToDie:
			fmt.Println("--- Break")
			break
		default:
			fmt.Println("--- Taking a photo")
			timer := time.NewTimer(time.Second * 2)
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
