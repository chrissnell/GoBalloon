// GoBalloon
// geospatial.go - Functions for APRS-related geospatial calculations
//
// (c) 2014, Christopher Snell

package geospatial

import (
	"fmt"
	"math"
	"time"
)

type Point struct {
	Lat            float64
	Lon            float64
	Altitude       float64
	Speed          float32
	Heading        uint16
	RadioRange     float32
	MessageCapable bool
	Time           time.Time
}

func NewPoint() *Point {
	return new(Point)
}

func (p *Point) DistanceTo(t *Point) (d uint32) {
	d = 1234
	return d
}

// APRS latitude format:  DDMM.mm
func LatDecimalDegreesToDegreesDecimalMinutes(d float64) string {
	deg := math.Floor(d)
	min := (d - deg) * 60
	ddm := fmt.Sprintf("%02d%02.2f", int(deg), min)
	return ddm
}

// APRS longitude format: DDDMM.mm
func LonDecimalDegreesToDegreesDecimalMinutes(d float64) string {
	deg := math.Floor(d)
	min := (d - deg) * 60
	ddm := fmt.Sprintf("%03d%02.2f", int(deg), min)
	return ddm
}
