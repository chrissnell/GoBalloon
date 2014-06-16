// GoBalloon
// encoder.go - AX.25/KISS encoder
//
// This code borrows heavily from Dustin Sallings's go-aprs library.
// https://github.com/dustin/go-aprs
// I've modified the code to play nicely with hardware TNCs using KISS and
// added comments that explain the structure of the AX.25/KISS packets

package ax25

import (
	"bytes"
	"errors"
)

// A mask of 11100000, merged into the SSID byte with inclusive OR
// If this is an AX.25 command packet, we merge it into the destination
// address.  Otherwise, for an AX.25 response packet, we merge it into the
// source address
var setSSIDMask = byte(0x70 << 1)

// A mask of 01100000, merged into the SSID byte with inclusive OR
// If this is an AX.25 command packet, we merge it into the source
// address.  Otherwise, for an AX.25 response packet, we merge it into the
// destination address
var clearSSIDMask = byte(0x30 << 1)

// This encodes an AX.25 command packet.  It is differentiated from
// the response packet function below by the bitmask applied to the SSID bytes.
func EncodeAX25Command(in APRSData) ([]byte, error) {
	return CreatePacket(in, clearSSIDMask, setSSIDMask)
}

func EncodeAX25Response(in APRSData) ([]byte, error) {
	return CreatePacket(in, setSSIDMask, clearSSIDMask)
}

func CreatePacket(a APRSData, smask, dmask byte) (em []byte, err error) {

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

	p := &bytes.Buffer{}

	// First, we send a Frame End (FEND)
	p.Write([]byte{0xc0})

	// Next comes our command field
	p.Write([]byte{0x00})

	// Next comes the destination address
	p.Write(encodeAX25Address(a.Dest, dmask))

	// Then the source address.  This part is a little tricky.
	// If we are going to use digipeaters (i.e. we have a path), we have to set
	// the last (least significant) bit of the SSID byte to 0.  If we *don't* have
	// a path, we set it to 1.
	mask := smask
	if len(a.Path) == 0 {
		// We don't have a path, so we set that last bit to 1
		mask |= 1
	}
	// Now that we have our mask, we can encode our source address
	p.Write(encodeAX25Address(a.Source, mask))

	// Then our digipeater path
	for i, v := range a.Path {

		mask = byte(0x70 << 1)

		// If this is the last station in the path, we also set our mask
		// 0x61.   I'm not totally sure on this part and can't square it
		// with the AX.25 spec but it works, so there.
		if i == len(a.Path)-1 {
			mask = byte(0x61)
		}
		p.Write(encodeAX25Address(v, mask))
	}

	// Then a control field (0x03 signifies that this is a UI-Frame)
	p.Write([]byte{0x03})

	// Then comes our protocol ID
	p.Write([]byte{0xf0})

	// Now comes the information field: the actual APRS data
	p.WriteString(a.Body)

	// Finally, another FEND
	p.Write([]byte{0xc0})

	return p.Bytes(), nil

}

func encodeAX25Address(in APRSAddress, mask byte) []byte {
	out := make([]byte, 7)

	for i := 0; i < len(out); i++ {
		out[i] = 0x40
	}

	for i, p := range in.Callsign {
		out[i] = byte(p) << 1
	}

	out[6] = mask | (byte(in.SSID) << 1)

	return out

}
