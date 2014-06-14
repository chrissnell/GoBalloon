package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"log"
	//	"os"
	//	"os/signal"
	// "github.com/chrissnell/go-base91"
)

func main() {

	psource := ax25.APRSAddress{
		Callsign: "VA7MPG",
		SSID:     14,
	}

	pdest := ax25.APRSAddress{
		Callsign: "T7QUUP",
		SSID:     0,
	}

	path1 := ax25.APRSAddress{
		Callsign: "W7PFR",
		SSID:     1,
	}
	path2 := ax25.APRSAddress{
		Callsign: "WIDE1",
		SSID:     0,
	}
	path3 := ax25.APRSAddress{
		Callsign: "WIDE2",
		SSID:     2,
	}

	a := ax25.APRSData{
		Source: psource,
		Dest:   pdest,
		Path:   []ax25.APRSAddress{path1, path2, path3},
		Body:   "`2;G rhk/]\"51}",
	}

	packet, err := ax25.EncodeAX25Command(a)
	if err != nil {
		log.Fatalf("Unable to create packet: %v", err)
	} else {
		fmt.Printf("--> %v\n", string(packet))
	}

	fmt.Println("Byte#\tHexVal\tChar\tChar>>1\tBinary")
	fmt.Println("-----\t------\t----\t-------\t------")
	for k, v := range packet {
		rs := v >> 1
		fmt.Printf("%4d \t%#x \t%v \t%v\t%08b\n", k, v, string(v), string(rs), v)
	}

}
