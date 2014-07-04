// GoBalloon
// parser.go - Functions for parsing APRS packets and dispatching the appropriate decoder
//
// (c) 2014, Christopher Snell

package aprs

import (
	_ "bytes"
	_ "errors"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"github.com/chrissnell/GoBalloon/geospatial"
	_ "strconv"
	_ "strings"
)

type APRSData struct {
	Position            geospatial.Point
	Message             Message
	StandardTelemetry   StdTelemetryReport
	CompressedTelemetry CompressedTelemetryReport
}

func ParsePacket(p *ax25.APRSPacket) *APRSData {
	ad := &APRSData{}
	d := []byte(p.Body)

	fmt.Printf("%v\n", p.Body)

	// Position reports are at least 14 chars long
	if len(d) >= 14 {
		// Position reports w/o timestamp start with ! or =
		if d[0] == byte('!') || d[0] == byte('=') {
			fmt.Println("---> Position packet (no timestamp)")
			// Compressed reports will have a symbol table ID in their second byte
			if d[1] == byte('/') || d[1] == byte('\\') {
				fmt.Println("Compressed position packet")
			}
		}
	}

	// Messages have colons at the 1st and 11th bytes
	if d[0] == ':' || d[10] == ':' {
		fmt.Println("---> Message packet")
	}

	return ad
}
