// Copyright 2016-present Liferay, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
