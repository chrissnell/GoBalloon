package ax25

import (
	"errors"
)

func CreatePacket(a APRSData) (em []byte, err error) {
	if len(a.Source.Callsign) < 4 {
		err = errors.New("Invalid source address.")
		return
	}

	if a.Body == "" {
		err = errors.New("APRS body is nil.")
		return
	}

	if a.Dest.Callsign == "" {
		a.Dest = APRSAddress{
			Callsign: "APZ001",
		}
	}
	destaddr := encodeAX25Address(a.Dest)

	sourceaddr := encodeAX25Address(a.Source)

	p := make([]byte, 1)

	// First field is the frame end (FEND)
	p = append(p, 0xc0)

	// Next comes our command field
	p = append(p, 0x00)

	// Next comes the destination address
	p = append(p, destaddr...)

	// Then the source address
	p = append(p, sourceaddr...)

	// Then our digipeater path
	for _, v := range a.Path {
		p = append(p, encodeAX25Address(v)...)
		p = append(p, byte(','))
	}

	// Then a control field (0x03 signifies that this is a UI-Frame)
	p = append(p, 0x03)

	// Then comes our protocol ID
	p = append(p, 0xf0)

	// Now comes the information field: the actual APRS data
	p = append(p, a.Body...)

	// Finally, another FEND
	p = append(p, 0xc0)

	return p, nil

}

func encodeAX25Address(in APRSAddress) []byte {
	out := make([]byte, 7)

	for i, p := range in.Callsign {
		out[i] = byte(p)
	}

	out[6] = byte(in.SSID)

	for i, p := range out {
		out[i] = p << 1
	}
	return out

}
