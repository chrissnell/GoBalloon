package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/geospatial"
)

func main() {
	p := geospatial.NewPoint()
	p.Lat = 24.910
	p.Lon = -114.301
	p.Alt = 2000
	fmt.Printf("point: %#v\n", p)

	p2 := geospatial.NewPoint()
	p2.Lat = 25.109
	p2.Lon = -112.121
	p2.Alt = 89
	fmt.Printf("point: %#v\n", p2)

	d := p.DistanceTo(p2)
	fmt.Printf("Distance from p to p2: %v\n", d)
}
