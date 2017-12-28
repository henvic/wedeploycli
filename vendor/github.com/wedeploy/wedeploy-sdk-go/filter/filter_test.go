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

package filter

import (
	"testing"

	"github.com/wedeploy/wedeploy-sdk-go/geo"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
	"github.com/wedeploy/wedeploy-sdk-go/qrange"
)

func TestAny(t *testing.T) {
	var want = `{
    "age": {
        "operator": "any",
        "value": [12, 21, 25]
    }
}`
	var got = Any("age", 12, 21, 25)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestBoundingBox(t *testing.T) {
	var want = `{
    "shape": {
        "operator": "gp",
        "value": [
            [20, 0],
            [0, 20]
        ]
    }
}`
	var got = BoundingBox("shape",
		geo.NewBoundingBox(geo.NewPoint(20, 0), geo.NewPoint(0, 20)))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestBoundingBoxByGeoPoints(t *testing.T) {
	var want = `{
    "xshape": {
        "operator": "gp",
        "value": [
            [20, 0],
            [0, 20]
        ]
    }
}`
	var got = BoundingBox("xshape", geo.NewPoint(20, 0), geo.NewPoint(0, 20))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestComposingAdd(t *testing.T) {
	var want = `{
  "and": [
    {
      "age": {
        "operator": ">",
        "value": 12
      }
    },
    {
      "age": {
        "operator": "<",
        "value": 15
      }
    }
  ]
}`
	var ageGt = Gt("age", 12)
	var got = ageGt.Add("and", "age", "<", 15)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestComposingAndFilters(t *testing.T) {
	var want = `{
    "and": [
        {
            "age": {
                "operator": ">",
                "value": 12
            }
        },
        {
            "age": {
                "operator": "<",
                "value": 15
            }
        },
        {
            "name": {
                "operator": "=",
                "value": "foo"
            }
        }
    ]
}`
	var got = And(New("age", ">", 12),
		New("age", "<", 15),
		Equal("name", "foo"),
	)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestComposingWithFilters(t *testing.T) {
	var want = `{
  "and": [
    {
      "age": {
        "operator": ">",
        "value": 12
      }
    },
    {
      "age": {
        "operator": "<",
        "value": 15
      }
    },
    {
      "city": {
        "operator": "=",
        "value": "Diamond Bar"
      }
    }
  ]
}`
	var ageGt = Gt("age", 12)
	var got = ageGt.Add("and", Lt("age", 15), Equal("city", "Diamond Bar"))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestComposingOrFilters(t *testing.T) {
	var want = `{
    "or": [
        {
            "age": {
                "operator": ">",
                "value": 12
            }
        },
        {
            "age": {
                "operator": "<",
                "value": 15
            }
        },
        {
            "name": {
                "operator": "=",
                "value": "foo"
            }
        }
    ]
}`
	var got = Or(New("age", ">", 12),
		New("age", "<", 15),
		Equal("name", "foo"),
	)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestDistanceCircle(t *testing.T) {
	var want = `{
    "point": {
        "operator": "gd",
        "value": {
            "location": [0, 0],
            "max": "2km"
        }
    }
}`
	var got = Distance("point", geo.NewCircle(geo.NewPoint(0, 0), "2km"), nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestDistancePointFrom(t *testing.T) {
	var want = `{
    "point": {
        "operator": "gd",
        "value": {
            "location": [0, 0],
            "min": 1
        }
    }
}`
	var got = Distance("point", geo.NewPoint(0, 0), qrange.From(1))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestDistancePointToImplicit(t *testing.T) {
	var want = `{
    "point": {
        "operator": "gd",
        "value": {
            "location": [0, 0],
            "max": 2
        }
    }
}`
	var got = Distance("point", geo.NewPoint(0, 0), 2)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestDistancePointTo(t *testing.T) {
	var want = `{
    "point": {
        "operator": "gd",
        "value": {
            "location": [0, 0],
            "max": 2
        }
    }
}`
	var got = Distance("point", geo.NewPoint(0, 0), qrange.To(2))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestDistancePointRange(t *testing.T) {
	var want = `{
    "point": {
        "operator": "gd",
        "value": {
            "location": [0, 0],
            "min": 1,
            "max": 2
        }
    }
}`
	var got = Distance("point", geo.NewPoint(0, 0), qrange.Between(1, 2))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestExists(t *testing.T) {
	var want = `{
    "age": {
        "operator": "exists"
    }
}`
	var got = Exists("age")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestEqualValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": "=",
        "value": 12
    }
}`
	var got = Equal("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFuzyFieldAndQuery(t *testing.T) {
	var want = `{
    "name": {
        "operator": "fuzzy",
        "value": {
            "query": "foo"
        }
    }
}`
	var got = Fuzzy("name", "foo", nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFuzyFieldAndFuziness(t *testing.T) {
	var want = `{
    "*": {
        "operator": "fuzzy",
        "value": {
            "query": "foo",
            "fuzziness": 0.8
        }
    }
}`
	var got = Fuzzy("foo", nil, 0.8)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFuzyQuery(t *testing.T) {
	var want = `{
    "*": {
        "operator": "fuzzy",
        "value": {
            "query": "foo"
        }
    }
}`
	var got = Fuzzy("foo", nil, nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestGreaterValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": ">",
        "value": 12
    }
}`
	var got = Gt("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestGreaterOrEqualValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": ">=",
        "value": 12
    }
}`
	var got = Gte("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestLessThanValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": "<",
        "value": 12
    }
}`
	var got = Lt("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestLessOrEqualValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": "=<",
        "value": 12
    }
}`
	var got = Lte("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}
func TestMoreThanValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": ">",
        "value": 12
    }
}`

	var got = New("age", ">", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestNotEqualValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": "!=",
        "value": 12
    }
}`
	var got = NotEqual("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestNilValue(t *testing.T) {
	var want = `{
    "age": {
        "operator": "="
    }
}`
	var got = Equal("age", nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestNone(t *testing.T) {
	var want = `{
    "age": {
        "operator": "none",
        "value": 12
    }
}`
	var got = None("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestNegation(t *testing.T) {
	var want = `{
    "not": {
        "age": {
            "operator": ">",
            "value": 12
        }
    }
}`
	var ageFilter = New("age", ">", 12)
	var got = ageFilter.Add("not")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestRegex(t *testing.T) {
	var want = `{
    "age": {
        "operator": "~",
        "value": 12
    }
}`
	var got = Regex("age", 12)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestMatch(t *testing.T) {
	var want = `{
    "xfield": {
        "operator": "match",
        "value": "foo"
    }
}`
	var got = Match("xfield", "foo")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestMatchAll(t *testing.T) {
	var want = `{"*":{"operator":"match","value":"foo"}}`
	var got = Match("foo")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestMissing(t *testing.T) {
	var want = `{
    "age": {
        "operator": "missing"
    }
}`
	var got = Missing("age")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestPhrase(t *testing.T) {
	var want = `{"*":{"operator":"phrase","value":"foo"}}`
	var got = Phrase("foo")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestPolygon(t *testing.T) {
	var want = `{
    "xshape": {
        "operator": "gp",
        "value": [
            [10, 0],
            [20, 0],
            [15, 10]
        ]
    }
}`
	var got = Polygon("xshape",
		geo.NewPoint(10, 0),
		geo.NewPoint(20, 0),
		geo.NewPoint(15, 10))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestPrefix(t *testing.T) {
	var want = `{
    "*": {
        "operator": "prefix",
        "value": "myPrefix"
    }
}`
	var got = Prefix("myPrefix")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestRange(t *testing.T) {
	var want = `{
    "age": {
        "operator": "range",
        "value": {
            "from": 12,
            "to": 15
        }
    }
}`
	var got = Range("age", 12, 15)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestRangeBetween(t *testing.T) {
	var want = `{
    "age": {
        "operator": "range",
        "value": {
            "from": 12,
            "to": 15
        }
    }
}`
	var got = Range("age", qrange.Between(12, 15))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestShapes(t *testing.T) {
	var want = `{
    "xshape": {
        "operator": "gs",
        "value": {
            "type": "geometrycollection",
            "geometries": [
                {
                    "type": "circle",
                    "coordinates": [0, 0],
                    "radius": "2km"
                },
                {
                    "type": "envelope",
                    "coordinates": [
                        [20, 0],
                        [0, 20]
                    ]
                }
            ]
        }
    }
}`

	var got = Shape("xshape", geo.NewCircle(geo.NewPoint(0, 0), "2km"),
		geo.NewBoundingBox(geo.NewPoint(20, 0), geo.NewPoint(0, 20)))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSimilarQuery(t *testing.T) {
	var want = `{
    "*": {
        "operator": "similar",
        "value": {
            "query": "foo"
        }
    }
}`
	var got = Similar("foo", nil, nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSimilarFieldAndQuery(t *testing.T) {
	var want = `{
    "name": {
        "operator": "similar",
        "value": {
            "query": "foo"
        }
    }
}`
	var got = Similar("name", "foo", nil)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSimilarFieldAndFuziness(t *testing.T) {
	var want = `{
    "*": {
        "operator": "similar",
        "value": {
            "query": "foo",
            "fuzziness": 0.8
        }
    }
}`
	var got = Similar("foo", nil, 0.8)
	jsonlib.AssertJSONMarshal(t, want, got)
}
