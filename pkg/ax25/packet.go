// GoBalloon
// packet.go - AX.25/APRS packet-related functions and data types

package ax25

import (
	"fmt"
)

// APRSPacket describes an AX.25 packet, as used in APRS land
type APRSPacket struct {
	Original     string
	Source       APRSAddress
	Dest         APRSAddress
	Path         []APRSAddress
	Body         string
	OriginalBody string
}

// APRSAddress represents an AX.25 source or destination address
type APRSAddress struct {
	Callsign string
	SSID     uint8
}

// Returns a string representation of a full AX.25 address
func (a APRSAddress) String() string {
	if a.SSID != 0 {
		return fmt.Sprintf("%s-%d", a.Callsign, a.SSID)
	}
	return a.Callsign
}
