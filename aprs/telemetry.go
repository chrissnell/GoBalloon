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

type TelemetryReport struct {
	A1 uint8
	A2 uint8
	A3 uint8
	A4 uint8
	A5 uint8
	D1 byte
}

func CreateTelemetryReport(r *TelemetryReport) string {
	var buffer bytes.Buffer

	var values []string

	// First byte in our telemetry report is the data type indicator.
	// The rune 'T' indicates a standard APRS telemetry report with
	// five analog values and one digital value
	buffer.WriteRune('T')

	// Next, we assemble the telemetry values into an slice of strings
	values = append(values, strconv.FormatUint(uint64(r.A1), 10), strconv.FormatUint(uint64(r.A2), 10),
		strconv.FormatUint(uint64(r.A3), 10), strconv.FormatUint(uint64(r.A4), 10),
		strconv.FormatUint(uint64(r.A5), 10), fmt.Sprintf("%08b", r.D1))

	buffer.WriteString(strings.Join(values, ","))

	return buffer.String()
}

func ParseTelemetryReport(s string) (r *TelemetryReport, err error) {
	var errMissingTelemElements = errors.New("Telemetry message has incorrect number of elements")
	var telems []string

	telems = strings.Split(s[1:], ",")
	if len(telems) != 6 {
		err = errMissingTelemElements
		return
	}

	a1, err := strconv.Atoi(telems[0])
	a2, err := strconv.Atoi(telems[1])
	a3, err := strconv.Atoi(telems[2])
	a4, err := strconv.Atoi(telems[3])
	a5, err := strconv.Atoi(telems[4])
	d1, err := strconv.ParseInt(telems[5], 2, 64)

	r = &TelemetryReport{A1: uint8(a1),
		A2: uint8(a2),
		A3: uint8(a3),
		A4: uint8(a4),
		A5: uint8(a5),
		D1: byte(d1)}

	return
}
