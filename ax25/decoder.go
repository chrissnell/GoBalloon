package ax25

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// APRSAddress represents an AX.25 source or destination address
type APRSAddress struct {
	Callsign string
	SSID     uint8
}

// AX.25 Information field
type Info string

type APRSData struct {
	Original string
	Source   APRSAddress
	Dest     APRSAddress
	Path     []APRSAddress
	Body     Info
}

type Decoder struct {
	r *bufio.Reader
}

const reasonableSize = 15

var errShortMsg = errors.New("Message unreasonably short")
var errTruncatedMsg = errors.New("Truncated message")

// NewDecoder gets a new decoder over this reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{bufio.NewReader(r)}
}

// Process the next APRS packet we get
func (d *Decoder) Next() (APRSData, error) {
	var err error
	frame := []byte{}
	// Keep reading so long as our frame is < 15 bytes
	for len(frame) <= reasonableSize {
		// Read forward until we encounter 0xc0 and return this data including that 0xc0.
		frame, err = d.r.ReadBytes(byte(0xc0))
		if err != nil {
			// Unable to read for some reason so we return an empty APRSData{} struct and our error
			return APRSData{}, err
		}
	}
	return decodeMessage(frame)
}

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

func decodeMessage(frame []byte) (dm APRSData, err error) {
	/*
	   KISS Packet format

	   KISS is an abbreviated form of AX.25 that's used for tranferring packet radio frames
	   from a TNC to software applications.  KISS provides an easy-to-parse format that's
	   generic and thus able to be supported by a wide range of amateur radio software.
	   In terms of APRS, the main distinction between a KISS frame an an AX.25 frame is the
	   different frame beginning/end delimiters and the lack of a frame checksum (FCS) in the
	   KISS frame.

	   A KISS frame looks something like this:
	   ------------------------------------------------------------------------------------
	   Frame End (FEND)   1 byte (0xc0)
	   Command            1 byte (0x00)
	   Dest Addr          7 bytes (Callsign + SSID, can be generic digipeater path)
	   Source Addr        7 bytes (Callsign + SSID)
	   Digipeater Addrs   0-56 bytes (Digipeater path)
	   Control field      1 byte (always 0x03; specifies that this is a UI-frame)
	   Protocol ID        1 byte (0xf0; no layer 3 protocol)
	   Information Field  1-256 bytes (APRS data, first char is APRS data type identifier)
	   Frame End (FEND)   1 byte (0xc0)
	   ------------------------------------------------------------------------------------

	   Source: TAPR APRS Specifiction  http://www.aprs.org/doc/APRS101.PDF
	           KISS Wikipedia page     http://en.wikipedia.org/wiki/KISS_(TNC)

	*/

	if len(frame) < reasonableSize {
		err = errShortMsg
		return
	}

	//origFrame := frame

	// Discard the first byte (0x00, the AX.25 Command field)
	frame = frame[:len(frame)-1]

	// Next comes the 7-byte destination address. AX.25 addresses are in the format CCCCCCS,
	// where C = callsign and S = SSID.  Since each btye of the address is shifted one
	// bit to the left, we'll use our decodeAddr() to decode it.  Gotta love 1980s protocols!
	dm.Dest = parseAX25Address(frame[1:8])

	// Next verse same as the first.  Same old protocol, could be worse.
	dm.Source = parseAX25Address(frame[8:15])

	// Initialize our message's path with an empty array of APRSAddress
	dm.Path = []APRSAddress{}

	// At this point, we can discard the parts of the packet we've already processed
	frame = frame[15:]

	// Now we're going to bite off 7-byte chunks of the frame and decode them as digipeater
	// addresses to be stored in our dm.Path array.  We stop when we reach the Control Field
	// section of packet, which is delimited by a 0x03.
	for len(frame) > 7 && frame[0] != 3 {
		dm.Path = append(dm.Path, parseAX25Address(frame[:7]))
		// As we parse each digipeater address, we remove it from the remaining frame
		frame = frame[7:]
	}

	// At this point, if there's less than 2 bytes remaining in the frame, or we don't
	// have the Control Field (0x03) or Protocol ID (0xf0) at the beginning of the remaining
	// frame, we have a truncated frame, so we throw an error.
	if len(frame) < 2 || frame[0] != 3 || frame[1] != 0xf0 {
		err = errTruncatedMsg
		return
	}

	// If we're still going now, we can safely assume that everything remaining is APRS
	// data.   We'll convert the remaining bytes to a string and save it as the APRSMessage body.
	dm.Body = Info(string(frame[2:]))

	return

}
