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

package aggregation

import (
	"testing"

	"github.com/wedeploy/wedeploy-sdk-go/geo"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
	"github.com/wedeploy/wedeploy-sdk-go/qrange"
)

func TestAverage(t *testing.T) {
	var aggregation = Avg("myName", "myField")
	var want = `{
    "myField": {
        "operator": "avg",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestCount(t *testing.T) {
	var aggregation = Count("myName", "myField")
	var want = `{
    "myField": {
        "operator": "count",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestDistance(t *testing.T) {
	var want = `{
    "myField": {
        "operator": "geoDistance",
        "name": "myName",
        "value": {
            "location": [0, 0],
            "ranges": [
                {
                    "from": 0
                },
                {
                    "to": 0
                }
            ]
        }
    }
}`
	var distance = Distance("myName",
		"myField",
		geo.NewPoint(0, 0),
		qrange.From(0),
		qrange.To(0))
	jsonlib.AssertJSONMarshal(t, want, distance)
}

func TestDistanceWithUnit(t *testing.T) {
	var want = `{
    "myField": {
        "operator": "geoDistance",
        "name": "myName",
        "value": {
            "location": [0, 0],
            "unit": "km",
            "ranges": [
                {
                    "from": 0
                },
                {
                    "to": 0
                }
            ]
        }
    }
}`
	var distance = Distance("myName",
		"myField",
		geo.NewPoint(0, 0),
		qrange.From(0),
		qrange.To(0))

	distance.Unit("km")

	jsonlib.AssertJSONMarshal(t, want, distance)
}

func TestDistanceComplex(t *testing.T) {
	var want = `{
    "myField": {
        "operator": "geoDistance",
        "name": "myName",
        "value": {
            "location": [0, 0],
            "unit": "km",
            "ranges": [
                {
                    "from": 0
                },
                {
                    "to": 0
                },
                {
                    "to": 0
                },
                {
                    "from": 0,
                    "to": 1
                },
                {
                    "from": 1
                }
            ]
        }
    }
}`
	var distance = Distance("myName",
		"myField",
		geo.NewPoint(0, 0),
		qrange.From(0),
		qrange.To(0))

	distance.Range(qrange.To(0))
	distance.Range(qrange.Between(0, 1))
	distance.Range(qrange.From(1)).Unit("km")

	jsonlib.AssertJSONMarshal(t, want, distance)
}

func TestDistanceComplex2(t *testing.T) {
	var want = `{
    "myField": {
        "operator": "geoDistance",
        "name": "myName",
        "value": {
            "location": [0, 0],
            "ranges": [
                {
                    "from": 0
                },
                {
                    "to": 0
                },
                {
                    "from": 0,
                    "to": 1
                }
            ]
        }
    }
}`
	var distance = Distance("myName",
		"myField",
		geo.NewPoint(0, 0),
		qrange.From(0),
		qrange.To(0))

	distance.Range(0, 1)

	jsonlib.AssertJSONMarshal(t, want, distance)
}
func TestExtendedStats(t *testing.T) {
	var aggregation = ExtendedStats("myName", "myField")
	var want = `{
    "myField": {
        "operator": "extendedStats",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestHistogram(t *testing.T) {
	var aggregation = Histogram("myName", "myField", 10)
	var want = `{
    "myField": {
        "operator": "histogram",
        "name": "myName",
        "value": 10
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestMax(t *testing.T) {
	var aggregation = Max("myName", "myField")
	var want = `{
    "myField": {
        "operator": "max",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestMin(t *testing.T) {
	var aggregation = Min("myName", "myField")
	var want = `{
    "myField": {
        "operator": "min",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestMissing(t *testing.T) {
	var aggregation = Missing("myName", "myField")
	var want = `{
    "myField": {
        "operator": "missing",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestNew(t *testing.T) {
	var aggregation = New("myName", "myField", "min", nil)
	var want = `{
    "myField": {
        "operator": "min",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestStats(t *testing.T) {
	var aggregation = Stats("myName", "myField")
	var want = `{
    "myField": {
        "operator": "stats",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestSum(t *testing.T) {
	var aggregation = Sum("myName", "myField")
	var want = `{
    "myField": {
        "operator": "sum",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}

func TestTerms(t *testing.T) {
	var aggregation = Terms("myName", "myField")
	var want = `{
    "myField": {
        "operator": "terms",
        "name": "myName"
    }
}`
	jsonlib.AssertJSONMarshal(t, want, aggregation)
}
