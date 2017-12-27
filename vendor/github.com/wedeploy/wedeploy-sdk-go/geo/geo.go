package geo

// BoundingBox geometric type
type BoundingBox struct {
	Type        string  `json:"type"`
	Coordinates []Point `json:"coordinates"`
}

// Circle geometric type
type Circle struct {
	Type        string `json:"type"`
	Coordinates Point  `json:"coordinates"`
	Radius      string `json:"radius"`
}

// Line geometric type
type Line struct {
	Type        string  `json:"type"`
	Coordinates []Point `json:"coordinates"`
}

// Point geometric type
type Point [2]float64

// Polygon geometric type
type Polygon struct {
	Type        string    `json:"type"`
	Coordinates [][]Point `json:"coordinates"`
}

// NewPoint creates a new point with the given geographic coordinates
func NewPoint(lat, lon float64) Point {
	var p Point

	p[0] = lat
	p[1] = lon

	return p
}

// NewBoundingBox creates a new bounding box
func NewBoundingBox(upperLeft, lowerRight Point) BoundingBox {
	return BoundingBox{
		Type:        "envelope",
		Coordinates: []Point{upperLeft, lowerRight},
	}
}

// NewCircle creates a new circle
func NewCircle(coordinates Point, radius string) Circle {
	return Circle{
		Type:        "circle",
		Coordinates: coordinates,
		Radius:      radius,
	}
}

// NewLine adds a new line formed by the given points
func NewLine(coordinates ...Point) Line {
	return Line{
		Type:        "linestring",
		Coordinates: coordinates,
	}
}

// NewPolygon creates a new polygon
func NewPolygon(coordinates ...Point) Polygon {
	var x = [][]Point{coordinates}

	return Polygon{
		Type:        "polygon",
		Coordinates: x,
	}
}

// AddHole adds a hole to the region of the polygon
func (p *Polygon) AddHole(coordinates ...Point) {
	p.Coordinates = append(p.Coordinates, coordinates)
}
