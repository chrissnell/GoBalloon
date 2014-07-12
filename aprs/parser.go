// GoBalloon
// parser.go - Functions for parsing APRS packets and dispatching the appropriate decoder
//
// (c) 2014, Christopher Snell

package aprs

import (
	_ "bytes"
	_ "errors"
	_ "fmt"
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
	SymbolTable         rune
	SymbolCode          rune
	Comment             string
}

func ParsePacket(p *ax25.APRSPacket) *APRSData {

	ad := &APRSData{}
	d := []byte(p.Body)

	// Position reports are at least 14 chars long
	if len(d) >= 14 {
		// Position reports w/o timestamp start with ! or =
		if d[0] == byte('!') || d[0] == byte('=') {
			// Compressed reports will have a symbol table ID in their second byte
			if d[1] == byte('/') || d[1] == byte('\\') {
				ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, _ = DecodeCompressedPositionReport(p.Body)
			} else {
				ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, _ = DecodeUncompressedPositionReportWithoutTimestamp(p.Body)
			}
		} else if d[0] == byte('/') || d[0] == byte('@') {
			// This looks like an uncompressed position report with a timestamp
			ad.Position, ad.SymbolTable, ad.SymbolCode, p.Body, _ = DecodeUncompressedPositionReportWithTimestamp(p.Body)
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
			ad.Message, p.Body, _ = DecodeMessage(p.Body)
		}
	}

	ad.Comment = p.Body
	return ad

}
