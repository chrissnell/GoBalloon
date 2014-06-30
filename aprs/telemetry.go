// GoBalloon
// telemetry.go - Functions for creating and decoding APRS telemetry reports
//
// (c) 2014, Christopher Snell

package aprs

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type StdTelemetryReport struct {
	A1      uint8
	A2      uint8
	A3      uint8
	A4      uint8
	A5      uint8
	Digital byte
}

type CompressedTelemetryReport struct {
	A1      uint16
	A2      uint16
	A3      uint16
	A4      uint16
	A5      uint16
	Digital byte
}

func CreateTelemetryReport(r *StdTelemetryReport) string {
	// First byte in our telemetry report is the data type indicator.
	// The rune 'T' indicates a standard APRS telemetry report with
	// five analog values and one digital value
	return fmt.Sprintf("T%03v,%03v,%03v,%03v,%03v,%08b", r.A1, r.A2, r.A3, r.A4, r.A5, r.Digital)
}

func CreateCompressedTelemetryReport(seq uint16, r *CompressedTelemetryReport) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteRune('|')

	seq = (seq + 1) & 0x1FFF

	sv, err := EncodeBase91Telemetry(seq)
	if err != nil {
		return "", err
	}
	buffer.Write(sv)

	A1e, err := EncodeBase91Telemetry(r.A1)
	if err != nil {
		return "", err
	}
	A2e, err := EncodeBase91Telemetry(r.A2)
	if err != nil {
		return "", err
	}
	A3e, err := EncodeBase91Telemetry(r.A3)
	if err != nil {
		return "", err
	}

	A4e, err := EncodeBase91Telemetry(r.A4)
	if err != nil {
		return "", err
	}

	A5e, err := EncodeBase91Telemetry(r.A5)
	if err != nil {
		return "", err
	}

	buffer.Write(A1e)
	buffer.Write(A2e)
	buffer.Write(A3e)
	buffer.Write(A4e)
	buffer.Write(A5e)

	return buffer.String(), nil

}

func ParseTelemetryReport(s string) (*StdTelemetryReport, error) {
	var err error
	var errMissingTelemElements = errors.New("Telemetry message has incorrect number of elements")
	var telems []string

	r := &StdTelemetryReport{}

	telems = strings.Split(s[1:], ",")
	if len(telems) != 6 {
		err = errMissingTelemElements
		return nil, err
	}

	a1, err := strconv.Atoi(telems[0])
	r.A1 = uint8(a1)
	a2, err := strconv.Atoi(telems[1])
	r.A2 = uint8(a2)
	a3, err := strconv.Atoi(telems[2])
	r.A3 = uint8(a3)
	a4, err := strconv.Atoi(telems[3])
	r.A4 = uint8(a4)
	a5, err := strconv.Atoi(telems[4])
	r.A5 = uint8(a5)

	di, err := strconv.Atoi(telems[5])
	r.Digital = byte(di)

	return r, err
}
