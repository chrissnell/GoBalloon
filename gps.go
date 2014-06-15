// GoBalloon
// gps.go - GPS controller code
//
// (c) 2014, Christopher Snell
// This code borrows heavily from https://github.com/stratoberry/go-gpsd
// Some portions (c) 2013 Stratoberry Pi Project

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Session struct {
	socket net.Conn
	reader *bufio.Reader
}

type GPSDSentence struct {
	Class string `json:"class"`
}

type TPVSentence struct {
	Class  string    `json:"class"`
	Tag    string    `json:"tag"`
	Device string    `json:"device"`
	Mode   int       `json:"mode"`
	Time   time.Time `json:"time"`
	Ept    float64   `json:"ept"`
	Lat    float64   `json:"lat"`
	Lon    float64   `json:"lon"`
	Alt    float64   `json:"alt"`
	Epx    float64   `json:"epx"`
	Epy    float64   `json:"epy"`
	Epv    float64   `json:"ev"`
	Track  float64   `json:"track"`
	Speed  float64   `json:"speed"`
	Climb  float64   `json:"climb"`
	Epd    float64   `json:"epd"`
	Eps    float64   `json:"eps"`
	Epc    float64   `json:"epc"`
}

func readFromGPSD(msg chan string) {
	session := new(Session)

	for {
		fmt.Println("--- Connecting to gpsd")
		session = new(Session)
		fmt.Println("--- Created new session")
		var err error
		session.socket, err = net.Dial("tcp", "127.0.0.1:2947")
		if err != nil {
			fmt.Printf("--- %v\n", err)
			fmt.Println("--- ERROR: Could not connect to gpsd.  Sleeping 5s and retrying.")
			time.Sleep(5000 * time.Millisecond)
			continue
		}

		_, err = session.socket.Write([]byte("?WATCH={\"enable\":true,\"json\":true}"))
		if err != nil {
			log.Printf("--- ERROR: Could not send WATCH command to gpsd: %v", err)
		}

		session.reader = bufio.NewReader(session.socket)

		for {
			line, err := session.reader.ReadString('\n')
			if err != nil {
				fmt.Println("--- ERROR: Could not read from GPSD. Sleeping 1s and retrying.")
				time.Sleep(1000 * time.Millisecond)
				break
			} else {
				msg <- line
			}
		}
	}
}

func processGPSDSentences(msg chan string) {
	var tpv *TPVSentence
	for {
		select {
		case m := <-msg:
			var classify GPSDSentence
			err := json.Unmarshal([]byte(m), &classify)
			if err != nil {
				log.Printf("--- ERROR: Could not unmarshal sentence %v\n", err)
				break
			}
			fmt.Println("--- Received a GPS sentence")
			if classify.Class == "TPV" {
				err := json.Unmarshal([]byte(m), &tpv)
				if err != nil {
					log.Printf("--- ERROR: Could not unmarshal TPV sentence: %v\n", err)
					break
				}
				fmt.Println("--- TPV sentence received")
				currentPosition.Lon = tpv.Lon
				currentPosition.Lat = tpv.Lat
				currentPosition.Alt = tpv.Alt
				fmt.Printf("--- LAT: %v   LON: %v   ALT: %v\n", tpv.Lat, tpv.Lon, tpv.Alt)
			}
		}
	}
}

func GPSRun() {
	msg := make(chan string, 100)

	go readFromGPSD(msg)
	go processGPSDSentences(msg)

}
