package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	RECV_BUFFER_LEN = 2048
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

func connectToGPSD() (session *Session) {
	fmt.Println("--- Connecting to gpsd")
	session = new(Session)
	var err error
	session.socket, err = net.Dial("tcp", "127.0.0.1:2947")
	if err != nil {
		fmt.Printf("--- %v\n", err)
		fmt.Println("--- ERROR: Could not connect to gpsd.  Sleeping 5s and retrying.")
		time.Sleep(5000 * time.Millisecond)
		connectToGPSD()
	}
	session.reader = bufio.NewReader(session.socket)
	return
}

func readFromGPSD(msg chan string, s *Session) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			fmt.Println("--- ERROR: Could not read from GPSD. Sleeping 1s and retrying.")
			time.Sleep(1000 * time.Millisecond)
		} else {
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
			fmt.Println("--- Received a GPS sentence")
			if classify.Class == "TPV" {
				err := json.Unmarshal([]byte(m), &tpv)
				if err != nil {
					log.Printf("--- ERROR: Could not unmarshal TPV sentence: %v\n", err)
					break
				}
				fmt.Println("--- TPV sentence received")
				fmt.Printf("--- LAT: %v   LON: %v\n", tpv.Lat, tpv.Lon)
			}
		}
	}
}

func GPSRun() {
	msg := make(chan string, 100)

	s := connectToGPSD()

	_, err := s.socket.Write([]byte("?WATCH={\"enable\":true,\"json\":true}"))
	if err != nil {
		log.Fatalln("--- ERROR: Could not send WATCH command to gpsd.")
	}

	go readFromGPSD(msg, s)
	go processGPSDSentences(msg)

}
