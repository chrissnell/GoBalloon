// GoBalloon
// parser.go - Functions for parsing APRS packets and dispatching the appropriate decoder
//
// (c) 2014, Christopher Snell

package aprs

import (
	"log"

	"github.com/chrissnell/GoBalloon/pkg/ax25"
	"github.com/chrissnell/GoBalloon/pkg/geospatial"
)

type APRSData struct {
	Position            geospatial.Point
	Message             Message
	StandardTelemetry   StdTelemetryReport
	CompressedTelemetry CompressedTelemetryReport
	SymbolTable         rune
	SymbolCode          rune
	Comment             string
}

func ParsePacket(p *ax25.APRSPacket) *APRSData {

	var err error

	ad := &APRSData{}
	d := []byte(p.Body)

	// Position reports are at least 14 chars long
	if len(d) >= 14 {
		// Position reports w/o timestamp start with ! or =
		if d[0] == byte('!') || d[0] == byte('=') {
			// Compressed reports will have a symbol table ID in their second byte
			if d[1] == byte('/') || d[1] == byte('\\') {
				ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, err = DecodeCompressedPositionReport(p.Body)
				if err != nil {
					log.Printf("Error decoding compressed position report: %v\n", err)
				}

			} else {
				ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, err = DecodeUncompressedPositionReportWithoutTimestamp(p.Body)
				if err != nil {
					log.Printf("Error decoding uncompressed position report without timestamp: %v\n", err)
				}

			}
		} else if d[0] == byte('/') || d[0] == byte('@') {
			// This looks like an uncompressed position report with a timestamp
			ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, err = DecodeUncompressedPositionReportWithTimestamp(p.Body)
			if err != nil {
				log.Printf("Error decoding uncompressed position report without timestamp: %v\n", err)
			}
		}
	}

	if len(d) >= 32 {
		// Signature of a standard uncompressed telemetry packet
		if d[0] == byte('T') && d[1] == byte('#') && d[5] == byte(',') {
			ad.StandardTelemetry, p.Body = ParseUncompressedTelemetryReport(p.Body)
		}
	}

	// Messages have colons at the 1st and 11th bytes
	if len(d) >= 11 {
		if d[0] == ':' || d[10] == ':' {
			ad.Message, p.Body, err = DecodeMessage(p.Body)
			if err != nil {
				log.Printf("Error decoding message: %v\n", err)
			}
			ad.Message.Sender = p.Source
		}
	}

	if len(d) >= 16 {
		ad.CompressedTelemetry, p.Body, err = ParseCompressedTelemetryReport(p.Body)
		if err != nil {
			log.Printf("Error decoding compressed telemetry report: %v\n", err)
		}

	}

	ad.Comment = p.Body
	return ad

}
