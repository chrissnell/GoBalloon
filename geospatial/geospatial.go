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

const (
	DegToRad = math.Pi / 180
	RadToDeg = 180 / math.Pi
)

func NewPoint() *Point {
	return new(Point)
}

func ToRadians(d float64) float64 {
	return d * DegToRad
}

func ToDegrees(d float64) float64 {
	return d * RadToDeg
}

func (p1 *Point) GreatCircleDistanceTo(p2 *Point) (d float64) {
	R := float64(6371)
	φ1 := ToRadians(p1.Lat)
	φ2 := ToRadians(p2.Lat)
	Δφ := ToRadians(p2.Lat - p1.Lat)
	Δλ := ToRadians(p2.Lon - p1.Lon)

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Return our GC distance in miles
	return R * c * 0.621371
}

func (p1 *Point) BearingTo(p2 *Point) uint16 {
	φ1 := ToRadians(p1.Lat)
	φ2 := ToRadians(p2.Lat)
	Δλ := ToRadians(p2.Lon - p1.Lon)

	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x)

	return ((uint16(ToDegrees(θ) + 360)) % 360)
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
