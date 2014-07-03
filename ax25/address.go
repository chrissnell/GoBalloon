// GoBalloon
// address.go - AX.25/APRS address-related functions and data types

package ax25

import (
	"fmt"
)

// APRSAddress represents an AX.25 source or destination address
type APRSAddress struct {
	Callsign string
	SSID     uint8
}

// Returns a string representation of a full AX.25 address
func (a APRSAddress) String() string {
	if a.SSID != 0 {
		return fmt.Sprintf("%s-%d", a.Callsign, a.SSID)
	} else {
		return a.Callsign
	}
}
