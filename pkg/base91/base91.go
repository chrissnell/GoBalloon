// GoBalloon
// base91.go - Functions for encoding and decoding APRS-style Base91 data
//
// (c) 2014, Christopher Snell

package base91

import (
	"bytes"
	"errors"
	"fmt"
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

func LatPrecompress(l float64) float64 {

	// Formula for pre-compression of latitude, prior to Base91 conversion
	p := 380926 * (90 - l)
	return p
}

func LonPrecompress(l float64) float64 {

	// Formula for pre-compression of longitude, prior to Base91 conversion
	p := 190463 * (180 + l)
	return p
}

func EncodeBase91Position(l int) []byte {
	b91 := make([]byte, 4)
	p1_div := int(l / (91 * 91 * 91))
	p1_rem := l % (91 * 91 * 91)
	p2_div := int(p1_rem / (91 * 91))
	p2_rem := p1_rem % (91 * 91)
	p3_div := int(p2_rem / 91)
	p3_rem := p2_rem % 91
	b91[0] = byte(p1_div) + 33
	b91[1] = byte(p2_div) + 33
	b91[2] = byte(p3_div) + 33
	b91[3] = byte(p3_rem) + 33
	return b91
}

func EncodeBase91Telemetry(l uint16) ([]byte, error) {

	if l > 8280 {
		err := errors.New("Cannot encode telemetry value larger than 8280")
		return nil, err
	}

	b91 := make([]byte, 2)
	p1_div := int(l / 91)
	p1_rem := l % 91
	b91[0] = byte(p1_div) + 33
	b91[1] = byte(p1_rem) + 33
	return b91, nil
}

func DecodeBase91Lat(p []byte) (float64, error) {
	if len(p) != 4 {
		return 0, fmt.Errorf("DecodeBase91Lat requires a four-byte slice as input.  Slice given: %v\n", p)
	}
	d := float64(90 - ((float64(p[0]-33))*(91*91*91)+(float64(p[1]-33))*(91*91)+(float64(p[2]-33))*91+float64(p[3]-33))/380926)

	return d, nil
}

func DecodeBase91Lon(p []byte) (float64, error) {
	if len(p) != 4 {
		return 0, fmt.Errorf("DecodeBase91Lot requires a four-byte slice as input.  Slice given: %v\n", p)
	}
	d := float64(-180 + ((float64(p[0]-33))*(91*91*91)+(float64(p[1]-33))*(91*91)+(float64(p[2]-33))*91+float64(p[3]-33))/190463)

	return d, nil
}

func DecodeBase91Altitude(p []byte) (float64, error) {
	if len(p) != 2 {
		return 0, fmt.Errorf("DecodeBase91Altitude requires a two-byte slice as input.  Slice given: %v\n", p)
	}
	cs := (float64(p[0]-33))*91 + float64(p[1]-33)
	alt := math.Pow(1.002, cs)

	return alt, nil
}

func DecodeBase91CourseSpeed(p []byte) (uint16, float32, error) {
	if len(p) != 2 {
		return 0, 0, fmt.Errorf("DecodeBase91CourseSpeed requires a two-byte slice as input.  Slice given: %v\n", p)
	}
	course := uint16((p[0] - 33) * 4)
	pow := float64(byte(p[1]) - 33)
	speed := float32(math.Pow(1.08, pow) - 1)
	return course, speed, nil
}

func DecodeBase91RadioRange(p byte) float32 {
	pow := float64(p - 33)
	rrange := float32(math.Pow(1.08, pow) * 2)
	return rrange
}

func DecodeBase91Telemetry(e []byte) (uint16, error) {
	if len(e) < 2 {
		return 0, fmt.Errorf("DecodeBase91Telemetry requires a two-byte slice as input.  Slice given: %v\n", e)
	}
	d := (int(e[0]-33))*91 + (int(e[1] - 33))
	return uint16(d), nil
}
