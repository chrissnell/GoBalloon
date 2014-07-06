package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/geospatial"
)

func main() {
	p := geospatial.NewPoint()
	p.Lat = 24.910
	p.Lon = -114.301
	p.Altitude = 10004
	fmt.Printf("point: %#v\n", p)

	position := aprs.CreateCompressedPositionReport(p, '/', 'O')
	fmt.Printf("Compressed position: %v\n", position)

	dp, st, sc, remains, err := aprs.DecodeCompressedPositionReport(position)

	fmt.Printf("Decoded compressed position: %+v\n", dp)
	fmt.Printf("symtable: %v   symcode: %v\n", st, sc)
	fmt.Printf("remains: %v\n", remains)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	u_pos, err := aprs.CreateUncompressedPositionReportWithoutTimestamp(p, '/', 'O', true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Uncompressed position: %v\n", u_pos)

	upos_dec, sym_t, sym_c, remains, err := aprs.DecodeUncompressedPositionReportWithoutTimestamp(u_pos)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Decoded uncompressed position:  %+v\n", upos_dec)
		fmt.Printf("symtable: %v   symcode: %v\n", sym_t, sym_c)
		fmt.Printf("remains: %v\n", remains)
	}
}
