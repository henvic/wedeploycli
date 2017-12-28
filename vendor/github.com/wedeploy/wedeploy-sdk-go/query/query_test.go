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

package query

import (
	"testing"

	"github.com/wedeploy/wedeploy-sdk-go/aggregation"
	"github.com/wedeploy/wedeploy-sdk-go/filter"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
)

func TestAggregate(t *testing.T) {
	var want = `{
    "aggregation": [
        {
            "f": {
                "operator": "min",
                "name": "a"
            }
        },
        {
            "f": {
                "operator": "missing",
                "name": "m"
            }
        }
    ]
}`
	var got = Aggregate("a", "f", "min")
	got.Aggregate(aggregation.Missing("m", "f"))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestAggregate2(t *testing.T) {
	var want = `{
    "aggregation": [
        {
            "bah": {
                "name": "foo"
            }
        }
    ]
}`
	var got = Aggregate("foo", "bah")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestCount(t *testing.T) {
	var want = `{"type":"count"}`
	var got = Count()
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFetch(t *testing.T) {
	var want = `{"type":"fetch"}`
	var got = Fetch()
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFilter(t *testing.T) {
	var want = `{
    "filter": [
        {
            "field": {
                "operator": "=",
                "value": 1
            }
        }
    ]
}`
	var got = Filter(filter.Equal("field", 1))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFilterString(t *testing.T) {
	var want = `{
    "filter": [
        {
            "field": {
                "operator": "=",
                "value": 1
            }
        }
    ]
}`
	var got = Filter("field", 1)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFilterThreeParams(t *testing.T) {
	var want = `{
    "filter": [
        {
            "age": {
                "operator": ">",
                "value": 18
            }
        }
    ]
}`
	var got = Filter("age", ">", 18)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestHighlight(t *testing.T) {
	var want = `{
    "highlight": [
        "field1",
        "field2",
        "field3"
    ]
}`
	var got = Highlight("field1").Highlight("field2").Highlight("field3")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestLimit(t *testing.T) {
	var want = `{"limit":1}`
	var got = Limit(1)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestLimitRewrite(t *testing.T) {
	var want = `{"limit":2}`
	var got = Limit(1).Limit(2)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestLimitAndOffset(t *testing.T) {
	var want = `{
    "limit": 1,
    "offset": 2
}`
	var got = Offset(2).Limit(1)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestOffset(t *testing.T) {
	var want = `{"offset":1}`
	var got = Offset(1)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestOffsetRewrite(t *testing.T) {
	var want = `{"offset":2}`
	var got = Offset(1).Offset(2)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearch(t *testing.T) {
	var want = `{
    "search": [
        {
            "xField": {
                "operator": "=",
                "value": "bah"
            }
        }
    ]
}`
	var got = Search(filter.Equal("xField", "bah"))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearchFieldQuery(t *testing.T) {
	var want = `{
    "search": [
        {
            "field": {
                "operator": "match",
                "value": "query"
            }
        }
    ]
}`
	var got = Search("field", "query")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearchFilterString(t *testing.T) {
	var want = `{
    "search": [
        {
            "field": {
                "operator": "=",
                "value": "value"
            }
        }
    ]
}`
	var got = Search(filter.Equal("field", "value"))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearchFieldOperatorQuery(t *testing.T) {
	var want = `{
    "search": [
        {
            "field": {
                "operator": "=",
                "value": "query"
            }
        }
    ]
}`
	var got = Search("field", "=", "query")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearchComplex(t *testing.T) {
	var want = `{
  "search": [
    {
      "*": {
        "operator": "match",
        "value": "query"
      }
    },
    {
      "*": {
        "operator": "match",
        "value": "query"
      }
    },
    {
      "field": {
        "operator": "match",
        "value": "value"
      }
    },
    {
      "field": {
        "operator": "=",
        "value": "value"
      }
    },
    {
      "field": {
        "operator": "=",
        "value": "value"
      }
    }
  ]
}
`
	var got = Search("query").Search("query")
	got.Search("field", "value").Search("field", "=", "value")
	got.Search(filter.Equal("field", "value"))
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSearchString(t *testing.T) {
	var want = `{
    "search": [
        {
            "*": {
                "operator": "match",
                "value": "query"
            }
        }
    ]
}`
	var got = Search("query")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSortDefault(t *testing.T) {
	var want = `{
    "sort": [
        {
            "field": "asc"
        }
    ]
}`
	var got = Sort("field")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSortMultiple(t *testing.T) {
	var want = `{
    "sort": [
        {
            "field1": "asc"
        },
        {
            "field2": "asc"
        },
        {
            "field3": "desc"
        }
    ]
}`
	var got = Sort("field1").Sort("field2", "asc").Sort("field3", "desc")
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestSortAndAggregate(t *testing.T) {
	var want = `{
    "sort": [
        {
            "field2": "asc"
        }
    ],
    "aggregation": [
        {
            "f": {
                "operator": "min",
                "name": "a"
            }
        },
        {
            "f": {
                "operator": "missing",
                "name": "m"
            }
        }
    ]
}`
	var got = Sort("field2").Aggregate("a", "f", "min")
	got.Aggregate(aggregation.Missing("m", "f"))
	jsonlib.AssertJSONMarshal(t, want, got)
}
