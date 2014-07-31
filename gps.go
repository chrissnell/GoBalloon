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
	"github.com/chrissnell/GoBalloon/geospatial"
	"log"
	"net"
	"sync"
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

type GPSReading struct {
	mu  sync.Mutex
	pos geospatial.Point
}

func (g *GPSReading) Set(pos geospatial.Point) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pos = pos
}

func (g *GPSReading) Get() geospatial.Point {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.pos
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

func processGPSDSentences(msg chan string, g *GPSReading) {
	var tpv *TPVSentence

	for {
		m := <-msg
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
			pos := geospatial.Point{Lon: tpv.Lon, Lat: tpv.Lat, Altitude: tpv.Alt * 3.28084, Speed: tpv.Speed, Heading: uint16(tpv.Track), Time: time.Now()}

			if *debug {
				log.Printf("Broadcasting position %+v\n", pos)
			}

			if pos.Lat != 0 {
				g.Set(pos)
			}

			if *debug {
				log.Printf("%+v\n", pos)
			}
		}
	}
}

func GPSRun(g *GPSReading) {
	msg := make(chan string)

	go readFromGPSD(msg)
	go processGPSDSentences(msg, g)

}
