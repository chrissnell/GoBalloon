// GoBalloon
// telemetry.go - Functions for creating and decoding APRS telemetry reports
//
// (c) 2014, Christopher Snell

package aprs

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type TelemetryReport struct {
	Analog  map[string]uint8
	Digital byte
}

func CreateTelemetryReport(r *TelemetryReport) string {
	var buffer bytes.Buffer

	tk := make([]string, len(r.Analog))

	i := 0
	for _, v := range r.Analog {
		if i < 5 {
			tk[i] = fmt.Sprintf("%03v", int(v))
		}
		i++
	}
	sort.Strings(tk)

	// First byte in our telemetry report is the data type indicator.
	// The rune 'T' indicates a standard APRS telemetry report with
	// five analog values and one digital value
	buffer.WriteRune('T')

	buffer.WriteString(strings.Join(tk, ","))

	buffer.WriteString(fmt.Sprintf(",%08b", r.Digital))

	return buffer.String()
}

func ParseTelemetryReport(s string) (*TelemetryReport, error) {
	var err error
	var errMissingTelemElements = errors.New("Telemetry message has incorrect number of elements")
	var telems []string

	r := &TelemetryReport{}

	tmap := make(map[string]uint8)

	telems = strings.Split(s[1:], ",")
	if len(telems) != 6 {
		err = errMissingTelemElements
		return nil, err
	}

	for i := 0; i < 5; i++ {
		ti, err := strconv.Atoi(telems[i])
		if err != nil {
			err = errors.New("Unable to convert telemetry value to unsigned integer (must be a whole number)")
			return nil, err
		}
		tmap[strconv.Itoa(i)] = uint8(ti)
	}

	r.Analog = tmap

	di, err := strconv.Atoi(telems[5])
	r.Digital = byte(di)

	return r, err
}
