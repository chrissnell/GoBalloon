// GoBalloon
// aprs.go - APRS controller code
//
// (c) 2014, Christopher Snell

package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/tarm/goserial"
	"io"
	"log"
	"net"
)

func connectToSerialTNC() (io.ReadWriteCloser, error) {

	fmt.Println("aprs_controller::connectToSerialTNC()")

	sc := &serial.Config{Name: *localtncport, Baud: 4800}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return s, fmt.Errorf("Error opening serial port %+v: %v\n", sc, err)
	}

	return s, nil

}

func connectToNetworkTNC() (io.ReadWriteCloser, error) {

	fmt.Println("aprs_controller::connectToNetworkTNC()")

	conn, err := net.Dial("tcp", *remotetnc)
	if err != nil {
		return io.ReadWriteCloser(conn), fmt.Errorf("Could not connect to %v.  Error: %v", *remotetnc, err)
	}
	return io.ReadWriteCloser(conn), nil
}

func incomingAPRSEventHandler(conn io.ReadWriteCloser) {

	fmt.Println("aprs_controller::incomingAPRSEventHandler()")

	d := ax25.NewDecoder(conn)

	defer conn.Close()

	for {

		// Retrieve a packet
		msg, err := d.Next()
		if err != nil {
			log.Printf("Error retrieving APRS message via KISS: %v", err)
		}

		fmt.Printf("Message received: %+v\n", msg)

		// Parse the packet
		ad := aprs.ParsePacket(&msg)

		if ad.Message.Recipient.Callsign == balloonAddr.Callsign && ad.Message.Recipient.SSID == balloonAddr.SSID {
			fmt.Printf("MESSAGE FOR ME: %+v\n", ad)
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

func StartAPRS() {

	var conn io.ReadWriteCloser
	var err error

	fmt.Println("aprs_controller::StartAPRS()")

	if len(*remotetnc) > 0 {
		conn, err = connectToNetworkTNC()
		if err != nil {
			log.Printf("Error connecting to TNC: %v\n", err)
		}
	} else if len(*localtncport) > 0 {
		conn, err = connectToSerialTNC()
		if err != nil {
			log.Printf("Error connecting to TNC: %v\n", err)
		}
	} else {
		log.Fatalln("Must provide either -remotetnc or -localtncport flag.")
	}

	go incomingAPRSEventHandler(conn)

}
