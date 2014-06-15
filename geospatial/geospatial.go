package geospatial

type Point struct {
	Lat float64
	Lon float64
	Alt float64
}

func NewPoint() *Point {
	return new(Point)
}

func (p *Point) DistanceTo(t *Point) (d uint32) {
	d = 1234
	return d
}
