// GoBalloon
// messaging.go - Functions for creating and decoding APRS messages
//
// (c) 2014-2018, Christopher Snell

package aprs

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chrissnell/GoBalloon/pkg/ax25"
)

// Message describes a station-to-station APRS message
type Message struct {
	Sender    ax25.APRSAddress
	Recipient ax25.APRSAddress
	ID        string
	Text      string
	ACK       bool // Set true if this is a message ACK response
	REJ       bool // Set true if this is a message REJ response
}

// EncodeMessage encodes Message into APRS message format
func EncodeMessage(m Message) (string, error) {
	var idtxt string

	if len(m.ID) != 0 {
		idtxt = "{" + string(m.ID)
	}
	return fmt.Sprintf(":%-9s:%s%s", m.Recipient.String(), m.Text, idtxt), nil
}

// EncodeMessageACK encodes a Message acknowledgement into APRS message ACK format
func EncodeMessageACK(m Message) (string, error) {

	if len(m.Sender.String()) == 0 {
		return "", errors.New("can't send an ACK without an addressee to reply to")
	}

	if len(m.ID) == 0 {
		return "", errors.New("can't send an ACK without a message ID to ACK")
	}

	return fmt.Sprintf(":%-9s:ack%s", m.Sender.String(), m.ID), nil
}

// DecodeMessage descodes a message in APRS message format into a Message
func DecodeMessage(m string) (Message, string, error) {
	var matches []string
	dm := Message{}

	if len(m) < 11 {
		return dm, m, fmt.Errorf("message length too short: should be >= 11 but is %v", len(m))
	}

	if m[0] != ':' || m[10] != ':' {
		return dm, m, errors.New("invalid message format.  1st and 10th characters should be ':'")
	}

	// APRS message regex from Hell.   Looks for the message, optional ACK/REJ, message ID, and whatever else.
	msgregex := regexp.MustCompile(`:([\w- ]{9}):([ackrejACKREJ]{3}[A-Za-z0-9]{1,5}$)?((.+)\{(\w{1,5}).*$)?(.*)$`)

	// If the message doesn't match the regex above *at all*, we store it in remains.
	// This kind of pattern allows us to run multiple parsers on an APRS packet and look
	// for valid APRS data with all of them.
	remains := msgregex.ReplaceAllString(m, "")

	if matches = msgregex.FindStringSubmatch(m); len(matches) > 0 {

		//
		if len(matches[6]) > 0 {
			remains = matches[6]
		}

		recipient := strings.TrimSpace(matches[1])

		if strings.Contains(recipient, "-") {
			rparts := strings.Split(recipient, "-")
			dm.Recipient.Callsign = rparts[0]
			ssid, err := strconv.ParseUint(rparts[1], 10, 8)
			if err != nil {
				return dm, remains, fmt.Errorf("error parsing SSID %v: %v", rparts[1], err)
			}
			dm.Recipient.SSID = uint8(ssid)
		} else {
			dm.Recipient.Callsign = recipient
		}

		if matches[2] != "" {
			if strings.ToLower(matches[2][0:3]) == "ack" {
				dm.ACK = true
				dm.ID = matches[2][3:]
				return dm, remains, nil
			}

			if strings.ToLower(matches[2][0:3]) == "rej" {
				dm.REJ = true
				dm.ID = matches[2][3:]
				return dm, remains, nil
			}
		}

		if matches[5] != "" {
			// This message has an ID so we capture it and don't include it with the message text
			dm.ID = matches[5]
			dm.Text = matches[4]
		} else {
			dm.Text = matches[6]
		}
		return dm, remains, nil

	}
	return dm, remains, nil

}
