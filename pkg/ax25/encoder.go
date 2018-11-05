// GoBalloon
// encoder.go - AX.25/KISS encoder
//
// Portions (c) 2014-2018 Chris Snell
//
// This code borrows from Dustin Sallings's go-aprs library.
// https://github.com/dustin/go-aprs
// I've modified the code to play nicely with hardware TNCs using KISS and
// added comments that explain the structure of the AX.25/KISS packets

package ax25

import (
	"bytes"
	"fmt"
)

// A mask of 11100000, to be OR'ed with the SSID byte.
// If this is an AX.25 command packet, we OR it with the destination
// address.  Otherwise, for an AX.25 response packet, we OR it with the
// source address
const setSSIDMask = byte(0x70 << 1)

// A mask of 01100000, to be OR'ed with the SSID byte.
// If this is an AX.25 command packet, we OR it with the source
// address.  Otherwise, for an AX.25 response packet, we OR it with the
// destination address
const clearSSIDMask = byte(0x30 << 1)

// EncodeAX25Command encodes an APRSPacket into an AX.25 command packet.
// It is differentiated from the EncodeAX25Response function by the bitmask
// that gets applied to the SSID bytes.
func EncodeAX25Command(in APRSPacket) ([]byte, error) {
	return CreatePacket(in, clearSSIDMask, setSSIDMask)
}

// EncodeAX25Response encodes an APRSPacket into an AX.25 response packet.
// It is differentiated from the EncodeAX25Command function by the bitmask
// that gets applied to the SSID bytes.
func EncodeAX25Response(in APRSPacket) ([]byte, error) {
	return CreatePacket(in, setSSIDMask, clearSSIDMask)
}

// CreatePacket creates an AX.25 packet from an APRSPacket and a pair
// of source and destination bitmasks, which are used to differentiate
// command packets from response packets.
func CreatePacket(a APRSPacket, smask, dmask byte) ([]byte, error) {

	if len(a.Source.Callsign) < 4 {
		return []byte{}, fmt.Errorf("invalid source address (length: %v bytes)", len(a.Source.Callsign))
	}

	if a.Body == "" {
		return []byte{}, fmt.Errorf("packet APRS body is nil")
	}

	// If a destination callsign was not provided, use the GoBalloon one
	if a.Dest.Callsign == "" {
		a.Dest = APRSAddress{
			Callsign: "APGBLN",
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
	if len(a.Path) == 0 {
		// We don't have a path, so we set that last bit to 1
		smask |= 1
	}
	// Now that we have our mask, we can encode our source address
	p.Write(encodeAX25Address(a.Source, smask))

	// Next, we encode and add each address of our digipeater path
	for i, addr := range a.Path {

		smask = byte(0x70 << 1)

		// If this is the last station in the path, we also set our mask
		// 0x61.   I'm not totally sure on this part and can't square it
		// with the AX.25 spec but it works, so there.
		if i == len(a.Path)-1 {
			smask = byte(0x61)
		}
		p.Write(encodeAX25Address(addr, smask))
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

// encodeAX25Address encodes an APRSAddress into the AX.25 format
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
