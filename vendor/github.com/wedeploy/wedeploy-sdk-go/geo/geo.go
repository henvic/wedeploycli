/*
Copyright (c) 2016-present, Liferay Inc. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of Liferay, Inc. nor the names of its contributors may
be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

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
