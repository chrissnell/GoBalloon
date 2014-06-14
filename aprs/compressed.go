package main

// var lat, lon float64 = 30.844, -72.75

// compressed_lat := int(LatPrecompress(lat))
// compressed_lon := int(LongPrecompress(lon))

// lat_byt := ToBase91(compressed_lat)
// lon_byt := ToBase91(compressed_lon)
// fmt.Println(lat_byt, lon_byt)
// fmt.Println(string(lat_byt[:]), string(lon_byt[:]))

func LatPrecompress(l float64) (p float64) {
	p = 380926 * (90 - l)
	return p
}

func LongPrecompress(l float64) (p float64) {
	p = 190463 * (180 + l)
	return p
}

func ToBase91(l int) (b91 [4]byte) {
	p1_div := int(l / (91 * 91 * 91))
	p1_rem := l % (91 * 91 * 91)
	p2_div := int(p1_rem / (91 * 91))
	p2_rem := p1_rem % (91 * 91)
	p3_div := int(p2_rem / 91)
	p3_rem := p2_rem % (91)
	b91[0] = byte(p1_div) + 33
	b91[1] = byte(p2_div) + 33
	b91[2] = byte(p3_div) + 33
	b91[3] = byte(p3_rem) + 33
	return b91
}
