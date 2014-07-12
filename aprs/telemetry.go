// GoBalloon
// telemetry.go - Functions for creating and decoding APRS telemetry reports
//
// (c) 2014, Christopher Snell

package aprs

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type StdTelemetryReport struct {
	Sequence uint16
	A1       float64
	A2       float64
	A3       float64
	A4       float64
	A5       float64
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

func CreateUncompressedTelemetryReport(r StdTelemetryReport) string {
	// First byte in our telemetry report is the data type indicator.
	// The rune 'T' indicates a standard APRS telemetry report with
	// five analog values and one digital value
	return fmt.Sprintf("T#%03v,%03v,%03v,%03v,%03v,%03v,%08b", r.Sequence, r.A1, r.A2, r.A3, r.A4, r.A5, r.Digital)
}

func CreateCompressedTelemetryReport(r CompressedTelemetryReport) (string, error) {
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

func ParseUncompressedTelemetryReport(s string) (StdTelemetryReport, string) {
	var matches []string

	r := StdTelemetryReport{}

	tr := regexp.MustCompile(`T#([\d.]{3}),([\d.]{3}),([\d.]{3}),([\d.]{3}),([\d.]{3}),([\d.]{3}),([01]{8})(.*)$`)

	remains := tr.ReplaceAllString(s, "")

	matches = tr.FindStringSubmatch(s)

	if matches = tr.FindStringSubmatch(s); len(matches) >= 6 {

		if len(matches[8]) > 0 {
			remains = matches[8]
		}

		seq, _ := strconv.ParseUint(matches[1], 10, 16)
		r.Sequence = uint16(seq)
		r.A1, _ = strconv.ParseFloat(matches[2], 64)
		r.A2, _ = strconv.ParseFloat(matches[3], 64)
		r.A3, _ = strconv.ParseFloat(matches[4], 64)
		r.A4, _ = strconv.ParseFloat(matches[5], 64)
		r.A5, _ = strconv.ParseFloat(matches[6], 64)

		r.Digital = convertBinaryStringToUint8(matches[7])
	}

	return r, remains
}

func convertBinaryStringToUint8(a string) byte {
	var b byte

	if a[0] == '1' {
		b |= 0x80
	}
	if a[1] == '1' {
		b |= 0x40
	}
	if a[2] == '1' {
		b |= 0x20
	}
	if a[3] == '1' {
		b |= 0x10
	}
	if a[4] == '1' {
		b |= 0x8
	}
	if a[5] == '1' {
		b |= 0x4
	}
	if a[6] == '1' {
		b |= 0x2
	}
	if a[7] == '1' {
		b |= 0x1
	}

	return b
}

func ParseCompressedTelemetryReport(s string) (CompressedTelemetryReport, string, error) {

	var err error

	r := CompressedTelemetryReport{}

	pr := regexp.MustCompile(`\|(..)(..)(..)(..)(..)(..)(..)\|(.*)$`)

	remains := pr.ReplaceAllString(s, "")

	if matches := pr.FindStringSubmatch(s); len(matches) > 0 {

		if len(matches[7]) > 0 {
			remains = matches[7]
		}

		fmt.Println("- - - - - - - - - - > Compressed Telemetry < - - - - - - - - - - -")

		r.Sequence, err = DecodeBase91Telemetry([]byte(matches[0]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		r.A1, err = DecodeBase91Telemetry([]byte(matches[1]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		r.A2, err = DecodeBase91Telemetry([]byte(matches[2]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		r.A3, err = DecodeBase91Telemetry([]byte(matches[3]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		r.A4, err = DecodeBase91Telemetry([]byte(matches[4]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		r.A5, err = DecodeBase91Telemetry([]byte(matches[5]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}

		dtm, err := DecodeBase91Telemetry([]byte(matches[6]))
		if err != nil {
			fmt.Printf("Error decoding Base91 telemetry: %v\n", err)
			return r, remains, err
		}
		r.Digital = byte(dtm)

	}

	return r, remains, nil
}
