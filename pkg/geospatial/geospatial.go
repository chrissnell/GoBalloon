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

const (
	degToRad = math.Pi / 180
	radToDeg = 180 / math.Pi
)

// Point describes the location, altitude, speed+heading, and APRS capability of an object
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

// NewPoint creates a new geospatial.Point
func NewPoint() *Point {
	return new(Point)
}

// ToRadians converts degrees to radians
func ToRadians(d float64) float64 {
	return d * degToRad
}

// ToDegrees converts radians to degrees
func ToDegrees(d float64) float64 {
	return d * radToDeg
}

// GreatCircleDistanceTo returns the great circle distance to another point for a Point object
func (p1 *Point) GreatCircleDistanceTo(p2 Point) (d float64) {
	// Formula from www.movable-type.co.uk/scripts/latlong.html
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

// BearingTo returns the bearing to another point for a Point object
func (p1 *Point) BearingTo(p2 Point) uint16 {
	// Formula from www.movable-type.co.uk/scripts/latlong.html
	φ1 := ToRadians(p1.Lat)
	φ2 := ToRadians(p2.Lat)
	Δλ := ToRadians(p2.Lon - p1.Lon)

	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x)

	return ((uint16(ToDegrees(θ) + 360)) % 360)
}

// LatDecimalDegreesToDegreesDecimalMinutes returns APRS-formatted latitude (DDMM.mm) for a decimal latitude
func LatDecimalDegreesToDegreesDecimalMinutes(d float64) string {
	deg := math.Floor(d)
	min := (d - deg) * 60
	ddm := fmt.Sprintf("%02d%02.2f", int(deg), min)
	return ddm
}

// LonDecimalDegreesToDegreesDecimalMinutes returns APRS-formatted longitude (DDDMM.mm) for a decimal longitude
func LonDecimalDegreesToDegreesDecimalMinutes(d float64) string {
	deg := math.Floor(d)
	min := (d - deg) * 60
	ddm := fmt.Sprintf("%03d%02.2f", int(deg), min)
	return ddm
}
