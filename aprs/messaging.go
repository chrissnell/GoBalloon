// GoBalloon
// messaging.go - Functions for creating and decoding APRS messages
//
// (c) 2014, Christopher Snell

package aprs

import (
	"errors"
	"fmt"
	"github.com/chrissnell/GoBalloon/ax25"
	"regexp"
	"strconv"
	"strings"
)

type Message struct {
	Sender    ax25.APRSAddress
	Recipient ax25.APRSAddress
	ID        string
	Text      string
	ACK       bool // Set true if this is a message ACK response
	REJ       bool // Set true if this is a message REJ response
}

func CreateMessage(m *Message) (string, error) {
	var idtxt string

	if len(m.ID) != 0 {
		idtxt = "{" + string(m.ID)
	}
	return fmt.Sprintf(":%-9s:%s%s", m.Recipient.String(), m.Text, idtxt), nil
}

func CreateMessageACK(m *Message) (string, error) {

	if len(m.Sender.String()) == 0 {
		return "", errors.New("Can't send an ACK without an addressee to reply to.")
	}

	if len(m.ID) == 0 {
		return "", errors.New("Can't send an ACK without a message ID to ACK.")
	}

	return fmt.Sprintf(":%-9s:ack%s", m.Sender.String(), m.ID), nil
}

func DecodeMessage(m string) (*Message, error) {
	dm := Message{}

	if len(m) < 11 {
		return &dm, fmt.Errorf("Message length too short.  Should be >= 11 but is %v.", len(m))
	}

	if m[0] != ':' || m[10] != ':' {
		return &dm, errors.New("Invalid message format.  1st and 10th characters should be ':'")
	}

	recipregex, _ := regexp.Compile(`:?(.{0,9}):?([a-zA-Z]{0,3})(.{0,5})`)
	r_field := recipregex.FindAllStringSubmatch(m, -1)[0][1]
	if len(r_field) != 9 {
		return &dm, fmt.Errorf("Message recipient has invalid length.  Should be 9 chars, space padded but is only %v.\n", len(r_field))
	}

	recipient := strings.TrimSpace(r_field)

	if strings.Contains(recipient, "-") {
		rparts := strings.Split(recipient, "-")
		dm.Recipient.Callsign = rparts[0]
		ssid, err := strconv.ParseUint(rparts[1], 10, 8)
		if err != nil {
			return &dm, fmt.Errorf("Error parsing SSID %v:", rparts[1], err)
		}
		dm.Recipient.SSID = uint8(ssid)
	} else {
		dm.Recipient.Callsign = recipient
	}

	if strings.ToLower(recipregex.FindAllStringSubmatch(m, -1)[0][2]) == "ack" {
		dm.ID = recipregex.FindAllStringSubmatch(m, -1)[0][3]
		dm.ACK = true
		return &dm, nil
	}

	if strings.ToLower(recipregex.FindAllStringSubmatch(m, -1)[0][2]) == "rej" {
		dm.ID = recipregex.FindAllStringSubmatch(m, -1)[0][3]
		dm.REJ = true
		return &dm, nil
	}

	textregex, _ := regexp.Compile(`:.{0,9}:(.*)`)
	dm.Text = textregex.FindAllStringSubmatch(m, -1)[0][1]

	return &dm, nil
}

// Decodes a message ACK/REJ and returns a Message object with the Recipient and ID fields populated
// along with a boolean (true for ACK, false for REJect)
// func DecodeMessageACK(m string) (*Message, bool, error) {
// 	dm := Message{}

// 	if len(m) < 14 {
// 		return &dm, false, fmt.Errorf("Message ACK length too short.  Should be >= 14 but is %v.", len(m))
// 	}

// 	var ackregex, _ = regexp.Compile(`:(.{9}):([a-z]{3})(.{1,5})`)
// 	recipient := strings.TrimSpace(ackregex.FindAllStringSubmatch(m, -1)[0][1])
// 	response := ackregex.FindAllStringSubmatch(m, -1)[0][2]
// 	dm.ID = ackregex.FindAllStringSubmatch(m, -1)[0][3]

// 	if strings.Contains(recipient, "-") {
// 		rparts := strings.Split(recipient, "-")
// 		dm.Recipient.Callsign = rparts[0]
// 		ssid, err := strconv.ParseUint(rparts[1], 10, 8)
// 		if err != nil {
// 			return &dm, false, fmt.Errorf("Error parsing SSID %v:", rparts[1], err)
// 		}
// 		dm.Recipient.SSID = uint8(ssid)
// 	} else {
// 		dm.Recipient.Callsign = recipient
// 	}

// 	return &dm, false, nil
// }
