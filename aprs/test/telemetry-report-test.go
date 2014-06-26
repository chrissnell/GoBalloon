package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"log"
)

func main() {
	r := aprs.TelemetryReport{A1: 10, A2: 20, A3: 30, A4: 40, A5: 50, D1: 77}
	tr := aprs.CreateTelemetryReport(&r)

	fmt.Printf("Telemetry report: %v\n", tr)

	p, err := aprs.ParseTelemetryReport(tr)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Printf("%v\n", p)
}
