// GoBalloon
// messaging.go - Functions for creating and decoding APRS messages
//
// (c) 2014, Christopher Snell

package aprs

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chrissnell/GoBalloon/pkg/ax25"
)

type Message struct {
	Sender    ax25.APRSAddress
	Recipient ax25.APRSAddress
	ID        string
	Text      string
	ACK       bool // Set true if this is a message ACK response
	REJ       bool // Set true if this is a message REJ response
}

func CreateMessage(m Message) (string, error) {
	var idtxt string

	if len(m.ID) != 0 {
		idtxt = "{" + string(m.ID)
	}
	return fmt.Sprintf(":%-9s:%s%s", m.Recipient.String(), m.Text, idtxt), nil
}

func CreateMessageACK(m Message) (string, error) {

	if len(m.Sender.String()) == 0 {
		return "", errors.New("Can't send an ACK without an addressee to reply to.")
	}

	if len(m.ID) == 0 {
		return "", errors.New("Can't send an ACK without a message ID to ACK.")
	}

	return fmt.Sprintf(":%-9s:ack%s", m.Sender.String(), m.ID), nil
}

func DecodeMessage(m string) (Message, string, error) {
	var matches []string
	dm := Message{}

	if len(m) < 11 {
		return dm, m, fmt.Errorf("Message length too short.  Should be >= 11 but is %v.", len(m))
	}

	if m[0] != ':' || m[10] != ':' {
		return dm, m, errors.New("Invalid message format.  1st and 10th characters should be ':'")
	}

	// APRS message regex from Hell.   Looks for the message, optional ACK/REJ, message ID, and whatever else.
	msgregex := regexp.MustCompile(`:([\w- ]{9}):([ackrejACKREJ]{3}[A-Za-z0-9]{1,5}$)?((.+)\{(\w{1,5}).*$)?(.*)$`)

	remains := msgregex.ReplaceAllString(m, "")

	if matches = msgregex.FindStringSubmatch(m); len(matches) > 0 {

		// For debugging odd messages...
		// i := 0
		// for _, v := range matches {
		// 	fmt.Printf("[%v] ---> %v\n", i, v)
		// 	i++
		// }

		if len(matches[6]) > 0 {
			remains = matches[6]
		}

		recipient := strings.TrimSpace(matches[1])

		if strings.Contains(recipient, "-") {
			rparts := strings.Split(recipient, "-")
			dm.Recipient.Callsign = rparts[0]
			ssid, err := strconv.ParseUint(rparts[1], 10, 8)
			if err != nil {
				return dm, remains, fmt.Errorf("Error parsing SSID %v:", rparts[1], err)
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

	} else {
		return dm, remains, nil
	}

}
