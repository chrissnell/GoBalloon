// GoBalloon
// decoder.go - AX.25/KISS decoder
//
// This code borrows from Dustin Sallings's go-aprs library.
// https://github.com/dustin/go-aprs
// I've modified the code to play nicely with hardware TNCs using KISS and
// added comments that explain the structure of the AX.25/KISS packets.

package ax25

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

// The minimum size (in bytes) for an APRS-style AX-25 packet
const minimumPacketSize = 15

// Decoder wraps a bufio.Reader and provides some methods for parsing
// AX.25 APRS packets that are read from it
type Decoder struct {
	r     *bufio.Reader
	debug bool
}

// NewDecoder wraps a Decoder around a provided Reader.
func NewDecoder(r io.Reader, debug bool) *Decoder {
	return &Decoder{
		r:     bufio.NewReader(r),
		debug: debug,
	}
}

// Next reads from our Reader and returns the next APRSPacket that it finds
func (d *Decoder) Next() (*APRSPacket, error) {
	var err error
	var frame []byte

	// Keep reading so long as our frame is < 15 bytes
	for len(frame) <= minimumPacketSize {
		// Read forward until we encounter 0xc0 and return this data including that 0xc0.
		frame, err = d.r.ReadBytes(byte(0xc0))
		if err != nil {
			return &APRSPacket{}, err
		}
	}

	if d.debug {
		log.Println("PACKET READ ---v")
		log.Println("Byte#\tHexVal\tChar\tChar>>1\tBinary")
		log.Println("Byte#\tHexVal\tChar\tChar>>1\tBinary")
		for k, v := range frame {
			rs := v >> 1
			fmt.Printf("%4d \t%#x \t%v \t%v\t%08b\n", k, v, string(v), string(rs), v)
		}
	}

	return decodePacket(frame)
}

// parseAX25Address parses an APRSAddress from a byte slice taken from our packet
func parseAX25Address(in []byte) APRSAddress {
	out := make([]byte, len(in))

	// We iterate through each byte of the address and shift right one bit.
	for i, p := range in {
		out[i] = p >> 1
	}

	// AX.25 addresses may be *up to* 7 bytes (6 bytes callsign, 1 byte SSID), but
	// they can be shorter.  Shorter callsigns are padded with spaces so we'll trim
	// them off.
	a := APRSAddress{
		Callsign: strings.TrimSpace(string(out[:len(out)-1])),
		SSID:     uint8(out[len(out)-1] & 0xf),
	}

	return a
}

// decodePacket decodes an APRSPacket from an APRS-style AX.25 packet
//
//
//    KISS Packet format
//
//    KISS is an abbreviated form of AX.25 that's used for tranferring packet radio frames
//    from a TNC to software applications.  KISS provides an easy-to-parse format that's
//    generic and thus able to be supported by a wide range of amateur radio software.
//    In terms of APRS, the main distinction between a KISS frame an an AX.25 frame is the
//    different frame beginning/end delimiters and the lack of a frame checksum (FCS) in the
//    KISS frame.
//
//    A KISS frame looks something like this:
//    ------------------------------------------------------------------------------------
//    Frame End (FEND)   1 byte (0xc0)
//    Command            1 byte (0x00)
//    Dest Addr          7 bytes (Callsign + SSID, can be generic digipeater path)
//    Source Addr        7 bytes (Callsign + SSID)
//    Digipeater Addrs   0-56 bytes (Digipeater path)
//    Control field      1 byte (always 0x03; specifies that this is a UI-frame)
//    Protocol ID        1 byte (0xf0; no layer 3 protocol)
//    Information Field  1-256 bytes (APRS data, first char is APRS data type identifier)
//    Frame End (FEND)   1 byte (0xc0)
//    ------------------------------------------------------------------------------------
//
//    Source: TAPR APRS Specifiction  						http://www.aprs.org/doc/APRS101.PDF
//            AX.25 Link-Layer Protocol Specification		https://www.tapr.org/pub_ax25.html
//            KISS Wikipedia page  						http://en.wikipedia.org/wiki/KISS_(TNC)
func decodePacket(frame []byte) (*APRSPacket, error) {
	packet := new(APRSPacket)

	if len(frame) < minimumPacketSize {
		return &APRSPacket{}, fmt.Errorf("Packet was too short (%v bytes)", len(frame))
	}

	// Discard the first byte (0x00, the AX.25 Command field)
	frame = frame[:len(frame)-1]

	// Next comes the 7-byte destination address. AX.25 addresses are in the format CCCCCCS,
	// where C = callsign and S = SSID.  Since each btye of the address is shifted one
	// bit to the left, we'll use our decodeAddr() to decode it.  Gotta love 1980s protocols!
	packet.Dest = parseAX25Address(frame[1:8])

	// Next verse same as the first.  Same old protocol, could be worse.
	packet.Source = parseAX25Address(frame[8:15])

	// At this point, we can discard the parts of the packet we've already processed
	frame = frame[15:]

	// Now we're going to bite off 7-byte chunks of the frame and decode them as digipeater
	// addresses to be stored in our path.  We stop when we reach the Control Field
	// section of packet, which is delimited by a 0x03.
	for len(frame) > 7 && frame[0] != 3 {
		packet.Path = append(packet.Path, parseAX25Address(frame[:7]))
		// As we parse each digipeater address, we remove it from the remaining frame
		frame = frame[7:]
	}

	// At this point, if there's less than 2 bytes remaining in the frame, or we don't
	// have the Control Field (0x03) or Protocol ID (0xf0) at the beginning of the remaining
	// frame, we have a truncated frame, so we throw an error.
	if len(frame) < 2 || frame[0] != 3 || frame[1] != 0xf0 {
		return &APRSPacket{}, fmt.Errorf("truncated frame detected; discarding packet")
	}

	// If we're still going now, we can safely assume that everything remaining is APRS
	// data, so we'll store it as our Body
	packet.Body = string(frame[2:])

	// APRSMessage.Body gets modified by the APRS decoder so we'll save a copy of it in the struct
	// that will remain in its original form
	packet.OriginalBody = packet.Body

	return packet, nil

}
