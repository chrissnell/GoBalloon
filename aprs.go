// GoBalloon
// aprs.go - APRS controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/chrissnell/GoBalloon/gps"
	"github.com/tarm/goserial"
)

type APRSTNC struct {
	conn            io.ReadWriteCloser
	netconn         net.Conn
	gps             *gps.GPSReading
	aprsPosition    chan geospatial.Point
	aprsMessage     chan string
	connecting      bool
	connectingMutex sync.Mutex
	connected       bool
	connectedMutex  sync.Mutex
	Remotetnc       *string
	Localtncport    *string
	Beaconint       *string
	symbolTable     rune
	symbolCode      rune
}

func (a *APRSTNC) IsConnected() bool {
	a.connectedMutex.Lock()
	defer a.connectedMutex.Unlock()
	return a.connected
}

func (a *APRSTNC) Connected(c bool) {
	a.connectedMutex.Lock()
	defer a.connectedMutex.Unlock()
	a.connected = c
}

func (a *APRSTNC) StartAPRS() {
	log.Println("APRSTNC.StartAPRS()")

	a.aprsMessage = make(chan string)
	a.aprsPosition = make(chan geospatial.Point)

	// We're going to block here until the TNC connection is established
	a.connectToTNC()

	go a.incomingAPRSEventHandler()
	go a.outgoingAPRSEventHandler()
	go a.StartAPRSPositionBeacon()

}

func (a *APRSTNC) connectToTNC() {
	if len(*a.Localtncport) > 0 {
		// Block on setting up a new connection to the serial TNC
		a.connectToSerialTNC()
	} else if len(*a.Remotetnc) > 0 {
		// Block on setting up a new connection to the serial TNC
		a.connectToNetworkTNC()
	} else {
		log.Fatalln("Must provide either -remotetnc or -localtncport flag.")
	}

}

func (a *APRSTNC) connectToSerialTNC() {
	var err error

	log.Println("APRSTNC.connectToSerialTNC()")

	// This mutex controls access to the boolean that indicates when a connect/reconnect
	// attempt is in progress
	a.connectingMutex.Lock()

	if a.connecting {
		a.connectingMutex.Unlock()
		log.Println("Skipping reconnect since a connection attempt is already in progress")
		return
	} else {
		// A connection attempt is not in progress so we'll start a new one
		a.connecting = true
		a.connectingMutex.Unlock()

		log.Println("Connecting to local TNC ", *a.Localtncport)

		for {
			sc := &serial.Config{Name: *a.Localtncport, Baud: 4800}
			a.conn, err = serial.OpenPort(sc)
			if err != nil {
				// There is a known problem where some shitty USB <-> serial adapters will drop out and Linux
				// will reattach them under a new device.  This code doesn't handle this situation currently
				// but it would be a nice enhancement in the future.
				log.Println("Sleeping 30 seconds and trying again")
				time.Sleep(30 * time.Second)
			} else {
				a.Connected(true)
				log.Printf("Connection to serial TNC on %v successful.\n", *a.Localtncport)

				a.connectingMutex.Lock()
				// Now that we've connected, we're no longer "connecting".  If a connection fails
				// and connectToSerialTNC() is called now, it should trigger a reconnect, so we
				// set a.connecting to false
				a.connecting = false
				a.connectingMutex.Unlock()
				return
			}
		}
	}
}

func (a *APRSTNC) connectToNetworkTNC() {
	var err error

	log.Println("APRSTNC.connectToNetworkTNC()")

	// This mutex controls access to the boolean that indicates when a connect/reconnect
	// attempt is in progress
	a.connectingMutex.Lock()

	if a.connecting {
		a.connectingMutex.Unlock()
		log.Println("Skipping reconnect since a connection attempt is already in progress")
		return
	} else {
		// A connection attempt is not in progress so we'll start a new one
		a.connecting = true
		a.connectingMutex.Unlock()

		log.Println("Connecting to remote TNC ", *a.Remotetnc)

		for {
			a.netconn, err = net.Dial("tcp", *a.Remotetnc)
			if err != nil {
				log.Printf("Could not connect to %v.  Error: %v", *a.Remotetnc, err)
				log.Println("Sleeping 5 seconds and trying again")
				time.Sleep(5 * time.Second)
			} else {
				a.Connected(true)
				log.Printf("Connection to TNC %v successful", a.netconn.RemoteAddr())
				a.conn = io.ReadWriteCloser(a.netconn)
				a.netconn.SetReadDeadline(time.Now().Add(time.Minute * 3))
				a.connectingMutex.Lock()
				// Now that we've connected, we're no longer "connecting".  If a connection fails
				// and connectToNetworkTNC() is called now, it should trigger a reconnect, so we
				// set a.connecting to false
				a.connecting = false
				a.connectingMutex.Unlock()
				return
			}
		}
	}
}

func (a *APRSTNC) incomingAPRSEventHandler() {

	log.Println("APRSTNC.incomingAPRSEventHandler()")

	for {

		// We loop the creation of this decoder so that it is recreated in the event that
		// the connection fails and we have to reconnect, creating a new a.conn and thus
		// necessitating a new Decoder over that new conn.
		d := ax25.NewDecoder(a.conn)

		for {

			// Retrieve a packet
			msg, err := d.Next()
			if err != nil {
				a.Connected(false)
				log.Printf("Error retrieving APRS message via KISS: %v", err)
				log.Println("Attempting to reconnect to TNC")
				// Reconnect to the TNC and break this inner loop so that a new Decoder
				// is created over the new connection
				a.connectToTNC()
				break
			}

			// Extend our read deadline on the net.Conn
			a.netconn.SetReadDeadline(time.Now().Add(time.Minute * 3))

			log.Printf("Incoming APRS packet received: %+v\n", msg)

			// Parse the packet
			ad := aprs.DecodePacket(&msg)

			// Look for messages addressed to the balloon
			if ad.Message.Recipient.Callsign == balloonAddr.Callsign && ad.Message.Recipient.SSID == balloonAddr.SSID {

				if strings.Contains(strings.ToUpper(ad.Message.Text), "CUTDOWN") {
					log.Println("CUTDOWN command received.  Initiating cutdown.")
					// Initiate cutdown When we receive the cutdown command
					InitiateCutdown()
				}

				// Send an ACK message in response to the cutdown command message
				ack, err := aprs.EncodeMessageACK(ad.Message)
				if err != nil {
					log.Printf("Error creating APRS message ACK: %v", err)
				}
				err = a.SendAPRSPacket(ack)
				if err != nil {
					log.Printf("Error sending APRS message ACK: %v", err)
				}
			}

		}

	}
}

func (a *APRSTNC) outgoingAPRSEventHandler() {

	var msg aprs.Message

	log.Println("APRSTNC.outgoingAPRSEventHandler()")

	for {
		select {
		case <-shutdownFlight:
			return

		case p := <-a.aprsPosition:

			// Send a postition packet
			pt := aprs.CreateCompressedPositionReport(p, a.symbolTable, a.symbolCode)

			log.Printf("Sending position report: %v\n", pt)
			err := a.SendAPRSPacket(pt)
			if err != nil {
				log.Printf("Error sending position report: %v\n", err)
			}

		case m := <-a.aprsMessage:

			msg.Recipient.Callsign = *chasercall
			ssidInt, _ := strconv.Atoi(*chaserssid)
			msg.Recipient.SSID = uint8(ssidInt)
			msg.Text = m
			msg.ID = "1"

			mt, err := aprs.EncodeMessage(msg)
			if err != nil {
				log.Printf("Error creating outgoing message: %v\n", err)
			}

			log.Printf("Sending message: %v\n", mt)
			err = a.SendAPRSPacket(mt)
			if err != nil {
				log.Printf("Error sending message: %v\n", err)
			}

		}
	}

}

func (a *APRSTNC) SendAPRSPacket(s string) error {

	var path []ax25.APRSAddress

	psource := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     7,
	}

	pdest := ax25.APRSAddress{
		Callsign: "APZ001",
		SSID:     0,
	}

	if a.gps.Get().Altitude > 3000 {
		path = append(path, ax25.APRSAddress{
			Callsign: "WIDE2",
			SSID:     1,
		})
	} else {
		path = append(path, ax25.APRSAddress{
			Callsign: "WIDE1",
			SSID:     1,
		})

		path = append(path, ax25.APRSAddress{
			Callsign: "WIDE2",
			SSID:     1,
		})
	}

	ap := ax25.APRSPacket{
		Source: psource,
		Dest:   pdest,
		Path:   path,
		Body:   s,
	}

	packet, err := ax25.EncodeAX25Command(ap)
	if err != nil {
		return fmt.Errorf("Unable to create packet: %v", err)
	}

	for {
		_, err = a.conn.Write(packet)
		if err != nil {
			a.Connected(false)
			log.Println("Error writing to TNC: ", err)
			log.Println("Attempting to reconnect to TNC")
			// Reconnect to the TNC and break this inner loop so that a new Decoder
			// is created over the new connection
			a.connectToTNC()
		} else {
			// Write was successful, so we break the loop
			break
		}
	}

	return nil

}

func (a *APRSTNC) StartAPRSPositionBeacon() {

	log.Println("APRSTNC.StartAPRSPositionBeacon()")

	for {
		p := a.gps.Get()
		log.Printf("Fetched new GPS point: %+v\n", p)
		if p.Lat != 0 && p.Lon != 0 {
			log.Printf("Sending APRS position for broadcast: %+v\n", p)
			a.aprsPosition <- p
		}
		interval, err := time.ParseDuration(fmt.Sprintf("%vs", *a.Beaconint))
		if err != nil {
			log.Fatalf("Invalid beacon interval.  Parsing error: %v\n", err)
		}
		time.Sleep(interval)
	}
}
