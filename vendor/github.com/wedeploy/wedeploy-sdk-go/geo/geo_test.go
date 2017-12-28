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
