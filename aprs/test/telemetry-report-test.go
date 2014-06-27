package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"log"
)

func main() {
	r := aprs.TelemetryReport{
		Analog:  map[string]uint8{"A": 77, "B": 10, "C": 20, "D": 30, "E": 40},
		Digital: 77,
	}
	tr := aprs.CreateTelemetryReport(&r)

	fmt.Printf("Telemetry report: %v\n", tr)

	p, err := aprs.ParseTelemetryReport(tr)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Printf("%v\n", p)
}
