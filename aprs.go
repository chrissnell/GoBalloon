// GoBalloon
// aprs.go - APRS controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	"github.com/tarm/goserial"
	"github.com/tv42/topic"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func connectToSerialTNC() (io.ReadWriteCloser, error) {

	log.Println("aprs_controller::connectToSerialTNC()")

	sc := &serial.Config{Name: *localtncport, Baud: 4800}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return s, fmt.Errorf("Error opening serial port %+v: %v\n", sc, err)
	}

	return s, nil

}

func connectToNetworkTNC() (io.ReadWriteCloser, error) {

	log.Println("aprs_controller::connectToNetworkTNC()")

	conn, err := net.Dial("tcp", *remotetnc)
	if err != nil {
		return io.ReadWriteCloser(conn), fmt.Errorf("Could not connect to %v.  Error: %v", *remotetnc, err)
	}
	return io.ReadWriteCloser(conn), nil
}

func incomingAPRSEventHandler(conn io.ReadWriteCloser) {

	log.Println("aprs_controller::incomingAPRSEventHandler()")

	d := ax25.NewDecoder(conn)

	for {

		// Retrieve a packet
		msg, err := d.Next()
		if err != nil {
			log.Printf("Error retrieving APRS message via KISS: %v", err)
		}

		log.Printf("Message received: %+v\n", msg)

		// Parse the packet
		ad := aprs.ParsePacket(&msg)

		// Look for messages addressed to the balloon
		if ad.Message.Recipient.Callsign == balloonAddr.Callsign && ad.Message.Recipient.SSID == balloonAddr.SSID {

			if strings.Contains(strings.ToUpper(ad.Message.Text), "CUTDOWN") {
				log.Println("CUTDOWN command received.  Initiating cutdown.")
				// Initiate cutdown When we receive the cutdown command
				InitiateCutdown()
			}

			// Send an ACK message in response to the cutdown command message
			ack, err := aprs.CreateMessageACK(ad.Message)
			if err != nil {
				log.Printf("Error creating APRS message ACK: %v", err)
			}
			err = SendAPRSPacket(ack, conn)
			if err != nil {
				log.Printf("Error sending APRS message ACK: %v", err)
			}
		}

	}
}

func outgoingAPRSEventHandler(conn io.ReadWriteCloser) {

	var msg aprs.Message

	log.Println("aprs_controller::outgoingAPRSEventHandler()")

	for {
		select {
		case <-shutdownFlight:
			return

		case p := <-aprsPosition:

			// Send a postition packet
			pt := aprs.CreateCompressedPositionReport(p, symbolTable, symbolCode)

			log.Printf("Sending position report: %v\n", pt)
			err := SendAPRSPacket(pt, conn)
			if err != nil {
				log.Printf("Error sending position report: %v\n", err)
			}

		case m := <-aprsMessage:

			msg.Recipient.Callsign = *chasercall
			ssidInt, _ := strconv.Atoi(*chaserssid)
			msg.Recipient.SSID = uint8(ssidInt)
			msg.Text = m
			msg.ID = "1"

			mt, err := aprs.CreateMessage(msg)
			if err != nil {
				log.Printf("Error creating outgoing message: %v\n", err)
			}

			log.Printf("Sending message: %v\n", mt)
			err = SendAPRSPacket(mt, conn)
			if err != nil {
				log.Printf("Error sending message: %v\n", err)
			}

		}
	}

}

func SendAPRSPacket(s string, conn io.ReadWriteCloser) error {

	var path []ax25.APRSAddress

	psource := ax25.APRSAddress{
		Callsign: "NW5W",
		SSID:     7,
	}

	pdest := ax25.APRSAddress{
		Callsign: "APZ001",
		SSID:     0,
	}

	if currentPosition.Altitude > 3000 {
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

	a := ax25.APRSPacket{
		Source: psource,
		Dest:   pdest,
		Path:   path,
		Body:   s,
	}

	packet, err := ax25.EncodeAX25Command(a)
	if err != nil {
		return fmt.Errorf("Unable to create packet: %v", err)
	}

	conn.Write(packet)

	return nil

}

func StartAPRSTNCConnector() {

	var conn io.ReadWriteCloser
	var err error

	log.Println("aprs_controller::StartAPRSTNCConnector()")

	for {
		if len(*remotetnc) > 0 {
			conn, err = connectToNetworkTNC()
			if err != nil {
				log.Printf("Error connecting to TNC: %v.  Sleeping 5sec and trying again.\n", err)
				time.Sleep(5 * time.Second)
				continue
			} else {
				break
			}
		} else if len(*localtncport) > 0 {
			conn, err = connectToSerialTNC()
			if err != nil {
				log.Printf("Error connecting to TNC: %v. Sleeping 5sec and trying again\n", err)
				time.Sleep(5 * time.Second)
				continue
			} else {
				break
			}
		} else {
			log.Fatalln("Must provide either -remotetnc or -localtncport flag.")
		}
	}

	go incomingAPRSEventHandler(conn)
	go outgoingAPRSEventHandler(conn)
}

func StartAPRSPositionBeacon(top *topic.Topic) {

	consumer := make(chan interface{}, 1)
	top.Register(consumer)

	defer top.Unregister(consumer)

	log.Println("aprs_controller::StartAPRSPositionBeacon()")

	for {
		p := <-consumer
		log.Println("Received new GPS point")
		if p != nil {
			pos := p.(geospatial.Point)
			if pos.Lat != 0 && pos.Lon != 0 {
				aprsPosition <- pos
			}
		}
		interval, err := time.ParseDuration(fmt.Sprintf("%vs", *beaconint))
		if err != nil {
			log.Fatalf("Invalid beacon interval.  Parsing error: %v\n", err)
		}
		time.Sleep(interval)
	}
}
