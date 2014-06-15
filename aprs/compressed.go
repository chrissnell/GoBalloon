package aprs

import (
	"bytes"
	"github.com/chrissnell/GoBalloon/geospatial"
	"math"
)

func CreateCompressedPosition(p *geospatial.Point, symTable, symCode rune) string {
	var buffer bytes.Buffer

	// First byte in our compressed position report is the data type indicator.
	// The rune '!' indicates a real-time compressed position report
	buffer.WriteRune('!')

	// Next byte is the symbol table selector
	buffer.WriteRune(symTable)

	// Next four bytes is the decimal latitude, compressed with funky Base91
	buffer.WriteString(string(positionToBase91(int(latPrecompress(p.Lat)))))

	// Then comes the longitude, same compression
	buffer.WriteString(string(positionToBase91(int(lonPrecompress(p.Lon)))))

	// Then our symbol code
	buffer.WriteRune(symCode)

	// Then we compress our altitude with a funky logrithm and conver to Base91
	buffer.Write(altitudeCompress(p.Alt))

	// This last byte specifies: a live GPS fix, in GGA NMEA format, with the
	// compressed position generated by software (this program!).  See APRS
	// Protocol Reference v1.0, page 39, for more details on this wack shit.
	buffer.WriteByte(byte(0x32) + 33)

	return buffer.String()
}

func altitudeCompress(a float64) []byte {
	var buffer bytes.Buffer

	// Altitude is compressed with the exponential equation:
	//   a = 1.002 ^ x
	//  where:
	//     a == altitude
	//     x == our pre-compressed altitude, to be converted to Base91
	precompAlt := int((math.Log(a) / math.Log(1.002)) + 0.5)

	// Convert our pre-compressed altitude to funky APRS-style Base91
	s := byte(precompAlt%91) + 33
	c := byte(precompAlt/91) + 33
	buffer.WriteByte(c)
	buffer.WriteByte(s)

	return buffer.Bytes()
}

func latPrecompress(l float64) (p float64) {

	// Formula for pre-compression of latitude, prior to Base91 conversion
	p = 380926 * (90 - l)
	return p
}

func lonPrecompress(l float64) (p float64) {

	// Formula for pre-compression of longitude, prior to Base91 conversion
	p = 190463 * (180 + l)
	return p
}

func positionToBase91(l int) (b91 []byte) {
	b91 = make([]byte, 4)
	p1_div := int(l / (91 * 91 * 91))
	p1_rem := l % (91 * 91 * 91)
	p2_div := int(p1_rem / (91 * 91))
	p2_rem := p1_rem % (91 * 91)
	p3_div := int(p2_rem / 91)
	p3_rem := p2_rem % (91)
	b91[0] = byte(p1_div) + 33
	b91[1] = byte(p2_div) + 33
	b91[2] = byte(p3_div) + 33
	b91[3] = byte(p3_rem) + 33
	return b91
}
