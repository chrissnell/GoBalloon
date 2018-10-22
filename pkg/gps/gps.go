// GoBalloon
// gps.go - GPS controller code
//
// (c) 2014-2018, Christopher Snell
// This borrows some NMEA parsing code from  https://github.com/stratoberry/go-gpsd
// Some portions (c) 2013 Stratoberry Pi Project

package gps

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"github.com/chrissnell/GoBalloon/pkg/geospatial"
)

// GPS holds our GPS configuration and runtime objects
type GPS struct {
	conn            net.Conn
	CurrentPosition CurrentPosition
	Remotegps       *string
	ready           bool
	readyMutex      sync.RWMutex
	debug           *bool
}

// Sentence is a generic sentence of uncertain type from gpsd
type Sentence struct {
	Class string `json:"class"`
}

// TPVSentence is a sentence of TPV type from gpsd
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

// CurrentPosition holds our current position as read from gpsd
type CurrentPosition struct {
	mu  sync.RWMutex
	pos geospatial.Point
}

// Set sets our current position to a provided Point
func (cp *CurrentPosition) Set(pos geospatial.Point) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.pos = pos
}

// Get fetches the current position
func (cp *CurrentPosition) Get() geospatial.Point {
	cp.mu.RLock()
	defer cp.mu.Unlock()
	return cp.pos
}

// IsReady returns whether the GPS is ready or not
func (g *GPS) IsReady() bool {
	g.readyMutex.RLock()
	defer g.readyMutex.Unlock()
	return g.ready
}

// Ready sets the current readiness state of the GPS
func (g *GPS) Ready(r bool) {
	g.readyMutex.Lock()
	defer g.readyMutex.Unlock()
	g.ready = r
}

// NewGPS creates a new connection to gpsd
func NewGPS(ctx context.Context, wg *sync.WaitGroup, remoteGPS string) *GPS {
	var g *GPS
	wg.Add(1)
	go g.StartGPS(ctx, wg)
	return g
}

// StartGPS connects to a gpsd server over TCP
func (g *GPS) StartGPS(ctx context.Context, wg *sync.WaitGroup) {
	var err error
	defer wg.Done()

	clientErr := make(chan error)

	for {
		select {
		case <-ctx.Done():
			return

		default:
			g.conn, err = net.Dial("tcp", *g.Remotegps)
			if err != nil {
				log.Printf("error connecting to %v: %v", *g.Remotegps, err)
				continue
			}

			wg.Add(1)
			go g.readFromGPSD(ctx, wg, clientErr)
			err = <-clientErr
		}
	}
}

func (g *GPS) readFromGPSD(ctx context.Context, wg *sync.WaitGroup, clientErr chan error) {
	var sentence Sentence
	var tpv *TPVSentence
	defer wg.Done()

	log.Println("GPS.incomingJSONHandler()")

	scanner := bufio.NewScanner(g.conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return

		default:
			// Read the line from our Scanner
			msg := scanner.Bytes()

			// Attempt to unmarshal it to a Sentence
			err := json.Unmarshal(msg, &sentence)
			if err != nil {
				clientErr <- err
				return
			}

			// We were able to read a JSON message, so we extend our read deadline
			g.conn.SetReadDeadline(time.Now().Add(time.Second * 15))

			switch sentence.Class {
			case "TPV":
				if err = json.Unmarshal(msg, &tpv); err != nil {
					clientErr <- err
					return
				}

				// Build our Point, converting altitude from meters to feet and speed from meters/sec to mph
				pos := geospatial.Point{Lon: tpv.Lon, Lat: tpv.Lat, Altitude: tpv.Alt * 3.28084, Speed: tpv.Speed * 2.236936, Heading: uint16(tpv.Track), Time: time.Now()}

				// Ensure a valid position
				if pos.Lat != 0 {
					if *g.debug {
						log.Printf("Saving position: %+v\n", pos)
					}

					// Set our position
					g.CurrentPosition.Set(pos)

					// Since we got a valid position, indicate that the GPS is ready
					g.Ready(true)
				} else {
					// The GPS is not ready
					g.Ready(false)
				}

			}
		}
	}

	if err := scanner.Err(); err != nil {
		clientErr <- err
		return
	}

}
