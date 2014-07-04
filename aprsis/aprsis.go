// GoBalloon
// aprsis.go - Provides an interface to the APRS-IS service
//
// This code comes from Dustin Sallings's go-aprs library:
// https://github.com/dustin/go-aprs
//
// I've tailored the code to suit my needs.

package aprsis

import (
	"errors"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/ax25"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"
)

var errEmptyMsg = errors.New("empty message")
var errInvalidMsg = errors.New("invalid message")

// An APRSIS connection.
type APRSIS struct {
	conn   *textproto.Conn
	rawLog io.Writer
}

// Next returns the next APRS message from this connection.
func (a *APRSIS) Next() (rv ax25.APRSPacket, err error) {
	var line string
	var dm *aprs.Message
	for err == nil || err == errEmptyMsg {
		line, err = a.conn.ReadLine()
		if err != nil {
			return
		}

		fmt.Fprintf(a.rawLog, "%s\n", line)

		if len(line) > 0 && line[0] != '#' {
			rv = ParseAPRSISPacket(line)

			dm, rv.Body, err = aprs.DecodeMessage(rv.Body)

			fmt.Printf("Message: %+v\n", dm)

			//if !rv.IsValid() {
			//	err = errInvalidMsg
			//}
			return rv, err
		}
	}

	return rv, errEmptyMsg
}

func ParseAPRSISPacket(i string) ax25.APRSPacket {

	parts := strings.SplitN(i, ":", 2)

	if len(parts) != 2 {
		return ax25.APRSPacket{}
	}
	srcparts := strings.SplitN(parts[0], ">", 2)
	if len(srcparts) < 2 {
		return ax25.APRSPacket{}
	}
	pathparts := strings.Split(srcparts[1], ",")

	return ax25.APRSPacket{Original: i,
		Source: AddressFromString(srcparts[0]),
		Dest:   AddressFromString(pathparts[0]),
		Path:   parseAddresses(pathparts[1:]),
		Body:   parts[1]}

}

// Parse addresses from the array of strings that we extraced from the packet's path
func parseAddresses(addrs []string) []ax25.APRSAddress {
	rv := []ax25.APRSAddress{}

	for _, s := range addrs {
		rv = append(rv, AddressFromString(s))
	}

	return rv
}

// AddressFromString builds an ax25.APRSAddress object from a string.
func AddressFromString(s string) ax25.APRSAddress {
	parts := strings.Split(s, "-")
	rv := ax25.APRSAddress{Callsign: parts[0]}
	if len(parts) > 1 {
		x, err := strconv.ParseInt(parts[1], 10, 32)
		if err == nil {
			rv.SSID = uint8(x)
		}
	}
	return rv
}

// SetRawLog sets a writer that will receive all raw APRS-IS messages.
func (a *APRSIS) SetRawLog(to io.Writer) {
	a.rawLog = to
}

// Dial an APRS-IS service.
func Dial(prot, addr string) (rv *APRSIS, err error) {
	var conn *textproto.Conn
	conn, err = textproto.Dial(prot, addr)
	if err != nil {
		return
	}

	return &APRSIS{
		conn:   conn,
		rawLog: ioutil.Discard,
	}, nil
}

// Auth authenticates and optionally set a filter.
func (a *APRSIS) Auth(user, pass, filter string) error {
	if filter != "" {
		filter = fmt.Sprintf(" filter %s", filter)
	}

	return a.conn.PrintfLine("user %s pass %s vers goaprs 0.1%s",
		user, pass, filter)
}
