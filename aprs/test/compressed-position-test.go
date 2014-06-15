package main

import (
	"fmt"
	"github.com/chrissnell/GoBalloon/aprs"
	"github.com/chrissnell/GoBalloon/geospatial"
)

func main() {
	p := geospatial.NewPoint()
	p.Lat = 24.910
	p.Lon = -114.301
	p.Alt = 10004
	fmt.Printf("point: %#v\n", p)

	position := aprs.CreateCompressedPosition(p, '/', 'O')
	fmt.Printf("Compressed position: %v\n", position)

}
