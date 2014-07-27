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
	Speed  float32   `json:"speed"`
	Climb  float64   `json:"climb"`
	Epd    float64   `json:"epd"`
	Eps    float64   `json:"eps"`
	Epc    float64   `json:"epc"`
}

func readFromGPSD(msg chan string) {
	session := new(Session)

	for {
		log.Println("--- Connecting to gpsd")
		session = new(Session)
		var err error
		session.socket, err = net.Dial("tcp", *remotegps)
		if err != nil {
			log.Printf("--- %v\n", err)
			log.Println("--- ERROR: Could not connect to gpsd.  Sleeping 5s and retrying.")
			time.Sleep(5000 * time.Millisecond)
			continue
		}

		_, err = session.socket.Write([]byte("?WATCH={\"enable\":true,\"json\":true}"))
		if err != nil {
			log.Printf("--- ERROR: Could not send WATCH command to gpsd: %v", err)
			continue
		}

		session.reader = bufio.NewReader(session.socket)

		lines := 0

		for {
			line, err := session.reader.ReadString('\n')
			lines += 1
			if lines > 100 {
				if *debug {
					log.Printf("%v lines received.  Disconnecting and reconnecting\n", lines)
				}
				break
			}
			if err != nil {
				log.Println("--- ERROR: Could not read from GPSD. Sleeping 1s and retrying.")
				time.Sleep(1000 * time.Millisecond)
				break
			}
			msg <- line
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
			if *debug {
				log.Println("--- Received a GPS sentence")
			}
			if classify.Class == "TPV" {
				err := json.Unmarshal([]byte(m), &tpv)
				if err != nil {
					log.Printf("--- ERROR: Could not unmarshal TPV sentence: %v\n", err)
					break
				}
				if *debug {
					log.Println("--- TPV sentence received")
				}
				currentPosition.Lon = tpv.Lon
				currentPosition.Lat = tpv.Lat
				currentPosition.Altitude = tpv.Alt * 3.28084 // meters to feet
				currentPosition.Speed = tpv.Speed
				currentPosition.Heading = uint16(tpv.Track)
				if *debug {
					log.Printf("--- LAT: %v   LON: %v   ALT: %v  SPD: %v   HDG: %v\n", tpv.Lat, tpv.Lon, tpv.Alt, tpv.Speed, tpv.Track)
				}
			}
		}
	}
}

func GPSRun() {
	msg := make(chan string, 100)

	go readFromGPSD(msg)
	go processGPSDSentences(msg)

}
