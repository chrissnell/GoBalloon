// GoBalloon
// base91.go - Functions for encoding and decoding APRS-style Base91 data
//
// (c) 2014, Christopher Snell

package aprs

import (
	"bytes"
	"math"
)

func AltitudeCompress(a float64) []byte {
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

func LatPrecompress(l float64) (p float64) {

	// Formula for pre-compression of latitude, prior to Base91 conversion
	p = 380926 * (90 - l)
	return p
}

func LonPrecompress(l float64) (p float64) {

	// Formula for pre-compression of longitude, prior to Base91 conversion
	p = 190463 * (180 + l)
	return p
}

func EncodeBase91Position(l int) (b91 []byte) {
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
