// GoBalloon
// gps.go - GPS controller code
//
// (c) 2014, Christopher Snell
// This code borrows heavily from https://github.com/stratoberry/go-gpsd
// Some portions (c) 2013 Stratoberry Pi Project

package gps

import (
	"bufio"
	"encoding/json"
	"github.com/chrissnell/GoBalloon/geospatial"
	"log"
	"net"
	"sync"
	"time"
)

type GPS struct {
	conn            net.Conn
	reader          *bufio.Reader
	Reading         GPSReading
	Remotegps       *string
	connecting      bool
	connectingMutex sync.Mutex
	ready           bool
	readyMutex      sync.Mutex
	msg             chan string
	Debug           *bool
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

func (gr *GPSReading) Set(pos geospatial.Point) {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	gr.pos = pos
}

func (gr *GPSReading) Get() geospatial.Point {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	return gr.pos
}

func (g *GPS) IsReady() bool {
	g.readyMutex.Lock()
	defer g.readyMutex.Unlock()
	return g.ready
}

func (g *GPS) Ready(r bool) {
	g.readyMutex.Lock()
	defer g.readyMutex.Unlock()
	g.ready = r
}

func (g *GPS) StartGPS() {
	log.Println("GPS.StartGPS()")

	g.msg = make(chan string)

	// Set up a new connection to the GPS
	g.connectToNetworkGPS()

	// Start a handler in a goroutine to read sentences off the Reader
	go g.incomingJSONHandler()

	// Start a processor in a goroutine to Unmarshal the JSON read by the handler
	go g.processJSONSentences()

}

func (g *GPS) connectToNetworkGPS() {
	var err error

	// This mutex controls access to the boolean that indicates when a connect/reconnect
	// attempt is in progress
	g.connectingMutex.Lock()

	if g.connecting {
		g.connectingMutex.Unlock()
		log.Println("Skipping reconnect since a connection attempt is already in progress")
		return
	} else {
		// A connection attempt is not in progress so we'll start a new one
		g.connecting = true
		g.connectingMutex.Unlock()

		log.Println("Connecting to remote GPS ", *g.Remotegps)

		for {
			g.conn, err = net.Dial("tcp", *g.Remotegps)
			if err != nil {
				log.Printf("Could not connect to %v.  Error: %v", *g.Remotegps, err)
				log.Println("Sleeping 5 seconds and trying again")
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("Connection to GPS %v successful", g.conn.RemoteAddr())
				g.conn.SetReadDeadline(time.Now().Add(time.Second * 15))

				_, err = g.conn.Write([]byte("?WATCH={\"enable\":true,\"json\":true}"))
				if err != nil {
					log.Println("Error sending WATCH command to GPS: ", err)
					log.Println("Attempting to reconnect to GPS")
					g.connectToNetworkGPS()
					continue
				}

				// Set up our reader
				g.reader = bufio.NewReader(g.conn)

				// We also need to declare that the GPS is ready
				g.Ready(true)

				g.connectingMutex.Lock()
				// Now that we've connected, we're no longer "connecting".  If a connection fails
				// and connectToNetworkGPS() is called now, it should trigger a reconnect, so we
				// set a.connecting to false
				g.connecting = false
				g.connectingMutex.Unlock()

				return
			}
		}
	}
}

func (g *GPS) incomingJSONHandler() {
	log.Println("GPS.incomingJSONHandler()")

	for {
		if g.IsReady() {
			line, err := g.reader.ReadString('\n')
			if err != nil {
				g.Ready(false)
				log.Printf("Error retrieving JSON message from GPS: %v", err)
				log.Println("Attempting to reconnect to the GPS")
				g.connectToNetworkGPS()
				continue
			}

			// Extend our read deadline
			g.conn.SetReadDeadline(time.Now().Add(time.Second * 15))

			// If we made it this far, we've successfully read a line so we send it over
			// the msg channel to be decoded elsewhere
			g.msg <- line
		}
	}
}

func (g *GPS) processJSONSentences() {
	var classify GPSDSentence
	var tpv *TPVSentence

	for {
		select {
		case m := <-g.msg:
			err := json.Unmarshal([]byte(m), &classify)
			if err != nil {
				log.Println("ERROR: Could not unmarshal sentence %v", err)
				continue
			}

			if *g.Debug {
				log.Printf("Received a GPS sentence: %v\n", m)
			}

			if classify.Class == "TPV" {
				err := json.Unmarshal([]byte(m), &tpv)
				if err != nil {
					log.Printf("ERROR: Could not unmarshal TPV sentence: %v\n", err)
					continue
				}

				if *g.Debug {
					log.Println("TPV sentence received")
				}

				// Build our Point, converting altitude from meters to feet and speed from meters/sec to mph
				pos := geospatial.Point{Lon: tpv.Lon, Lat: tpv.Lat, Altitude: tpv.Alt * 3.28084, Speed: tpv.Speed * 2.236936, Heading: uint16(tpv.Track), Time: time.Now()}

				if pos.Lat != 0 {
					if *g.Debug {
						log.Printf("Saving position: %v\n", pos)
					}

					g.Reading.Set(pos)
				}
			}
		}
	}
}
