package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"log"
)

func main() {
	s := aprs.StdTelemetryReport{
		A1:      206,
		A2:      100,
		A3:      20,
		A4:      36,
		A5:      40,
		Digital: 77,
	}

	c := aprs.CompressedTelemetryReport{
		A1:      7740,
		A2:      105,
		A3:      20,
		A4:      3006,
		A5:      403,
		Digital: 77,
	}

	tr := aprs.CreateTelemetryReport(&s)
	ctr, err := aprs.CreateCompressedTelemetryReport(901, &c)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	fmt.Printf("Telemetry report: %v\n", tr)
	fmt.Printf("Compressed Telemetry report: %v\n", ctr)

	p, err := aprs.ParseTelemetryReport(tr)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Printf("%v\n", p)
}
