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
	Sequence uint16
	A1       uint8
	A2       uint8
	A3       uint8
	A4       uint8
	A5       uint8
	Digital  byte
}

type CompressedTelemetryReport struct {
	Sequence uint16
	A1       uint16
	A2       uint16
	A3       uint16
	A4       uint16
	A5       uint16
	Digital  byte
}

func CreateUncompressedTelemetryReport(r *StdTelemetryReport) string {
	// First byte in our telemetry report is the data type indicator.
	// The rune 'T' indicates a standard APRS telemetry report with
	// five analog values and one digital value
	return fmt.Sprintf("T%03v,%03v,%03v,%03v,%03v,%03v,%08b", r.Sequence, r.A1, r.A2, r.A3, r.A4, r.A5, r.Digital)
}

func CreateCompressedTelemetryReport(r *CompressedTelemetryReport) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteRune('|')

	r.Sequence = (r.Sequence + 1) & 0x1FFF

	sv, err := EncodeBase91Telemetry(r.Sequence)
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

	if uint16(r.Digital) > 255 {
		err := errors.New("Digital value cannot exceed 8 bits (integer 255)")
		return "", err
	}

	D1e, err := EncodeBase91Telemetry(uint16(r.Digital))

	buffer.Write(A1e)
	buffer.Write(A2e)
	buffer.Write(A3e)
	buffer.Write(A4e)
	buffer.Write(A5e)
	buffer.Write(D1e)
	buffer.WriteRune('|')

	return buffer.String(), nil

}

func ParseUncompressedTelemetryReport(s string) (*StdTelemetryReport, error) {
	var err error
	var errMissingTelemElements = errors.New("Telemetry message has incorrect number of elements")
	var telems []string

	r := &StdTelemetryReport{}

	telems = strings.Split(s[1:], ",")
	if len(telems) != 7 {
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

func ParseCompressedTelemetryReport(s string) (*CompressedTelemetryReport, error) {

	var err error

	r := &CompressedTelemetryReport{}

	tbs := []byte(s)

	if len(s) != 16 {
		err := fmt.Errorf("Compressed telemetry message has incorrect length.  Should be 16, is %v.\n", len(s))
		return nil, err
	}

	r.Sequence, err = DecodeBase91Telemetry(tbs[1:3])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	r.A1, err = DecodeBase91Telemetry(tbs[3:5])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	r.A2, err = DecodeBase91Telemetry(tbs[5:7])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	r.A3, err = DecodeBase91Telemetry(tbs[7:9])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	r.A4, err = DecodeBase91Telemetry(tbs[9:11])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	r.A5, err = DecodeBase91Telemetry(tbs[11:13])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}

	dtm, err := DecodeBase91Telemetry(tbs[13:15])
	if err != nil {
		fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
		return nil, err
	}
	r.Digital = byte(dtm)

	return r, nil
}
