// Copyright 2016-present Liferay, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geo

import (
	"testing"

	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
)

func TestBoundingBox(t *testing.T) {
	var want = `{"type":"envelope","coordinates":[[0,20],[20,0]]}`
	var upperLeft = NewPoint(0, 20)
	var lowerRight = NewPoint(20, 0)
	var boundingBox = NewBoundingBox(upperLeft, lowerRight)
	jsonlib.AssertJSONMarshal(t, want, boundingBox)
}

func TestCircle(t *testing.T) {
	var want = `{"type":"circle","coordinates":[20,0],"radius":"2km"}`

	var coordinates = NewPoint(20, 0)

	var circle = NewCircle(coordinates, "2km")

	jsonlib.AssertJSONMarshal(t, want, circle)
}

func TestLine(t *testing.T) {
	var want = `{"type":"linestring","coordinates":[[10,20],[10,30],[10,40]]}`

	var line = NewLine(
		NewPoint(10, 20),
		NewPoint(10, 30),
		NewPoint(10, 40))

	jsonlib.AssertJSONMarshal(t, want, line)
}

func TestPoint(t *testing.T) {
	var point = NewPoint(10, 20)
	var want = "[10,20]"

	jsonlib.AssertJSONMarshal(t, want, point)
}

func TestPolygon(t *testing.T) {
	var want = `{
    "type": "polygon",
    "coordinates": [
        [
            [0,0],
            [0,30],
            [40,0]
        ],
        [
            [5,5],
            [5,8],
            [9,5]
        ]
    ]
}`

	var polygon = NewPolygon(
		NewPoint(0, 0),
		NewPoint(0, 30),
		NewPoint(40, 0))

	polygon.AddHole(
		NewPoint(5, 5),
		NewPoint(5, 8),
		NewPoint(9, 5))

	jsonlib.AssertJSONMarshal(t, want, polygon)
}
