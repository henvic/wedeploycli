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
	"github.com/wedeploy/wedeploy-sdk-go/aggregation"
	"github.com/wedeploy/wedeploy-sdk-go/filter"
)

// Builder is the query builder type
type Builder struct {
	Type        string                     `json:"type,omitempty"`
	Aggregation *[]aggregation.Aggregation `json:"aggregation,omitempty"`
	BFilter     *[]filter.Filter           `json:"filter,omitempty"`
	Highlights  *[]string                  `json:"highlight,omitempty"`
	BOffset     *int                       `json:"offset,omitempty"`
	BLimit      *int                       `json:"limit,omitempty"`
	BSearch     *[]filter.Filter           `json:"search,omitempty"`
	BSort       *[]map[string]string       `json:"sort,omitempty"`
}

// Aggregate creates a Aggregate query builder
func Aggregate(ai ...interface{}) *Builder {
	return New().Aggregate(ai...)
}

// Count creates a Count query builder
func Count() *Builder {
	return New().Count()
}

// Fetch creates a Fetch query builder
func Fetch() *Builder {
	return New().Fetch()
}

// Filter creates a Filter query builder
func Filter(ai ...interface{}) *Builder {
	return New().Filter(ai...)
}

// Highlight creates a Highlight query builder
func Highlight(field string) *Builder {
	return New().Highlight(field)
}

// Limit creates a Limit query builder
func Limit(limit int) *Builder {
	return New().Limit(limit)
}

// New creates a query builder
func New() *Builder {
	return &Builder{}
}

// Offset creates a Offset query builder
func Offset(offset int) *Builder {
	return New().Offset(offset)
}

// Search creates a Search query builder
func Search(ai ...interface{}) *Builder {
	return New().Search(ai...)
}

// Sort creates a Sort query builder
func Sort(field string, direction ...string) *Builder {
	return New().Sort(field, direction...)
}

// Aggregate adds new aggregations
func (b *Builder) Aggregate(ai ...interface{}) *Builder {
	var a *aggregation.Aggregation
	var operator interface{}

	switch len(ai) {
	case 1:
		a = ai[0].(*aggregation.Aggregation)
	default:
		if len(ai) == 3 {
			operator = ai[2]
		}

		a = aggregation.New(ai[0].(string), ai[1].(string), operator, nil)
	}

	if b.Aggregation == nil {
		b.Aggregation = &[]aggregation.Aggregation{}
	}

	*b.Aggregation = append(*b.Aggregation, *a)
	return b
}

// Count sets the query type to count
func (b *Builder) Count() *Builder {
	b.Type = "count"
	return b
}

// Fetch sets the query type to fetch
func (b *Builder) Fetch() *Builder {
	b.Type = "fetch"
	return b
}

// Filter adds new filters
func (b *Builder) Filter(ai ...interface{}) *Builder {
	var f filter.Filter

	switch len(ai) {
	case 1:
		f = *ai[0].(*filter.Filter)
	case 3:
		f = *filter.New(ai[0].(string), ai[1].(string), ai[2])
	default:
		f = *filter.Equal(ai[0].(string), ai[1])
	}

	if b.BFilter == nil {
		b.BFilter = &[]filter.Filter{}
	}

	*b.BFilter = append(*b.BFilter, f)

	return b
}

// Highlight field
func (b *Builder) Highlight(field string) *Builder {
	if b.Highlights == nil {
		b.Highlights = &[]string{}
	}

	*b.Highlights = append(*b.Highlights, field)

	return b
}

// Limit for the query
func (b *Builder) Limit(limit int) *Builder {
	b.BLimit = &limit
	return b
}

// Offset for the query
func (b *Builder) Offset(offset int) *Builder {
	b.BOffset = &offset
	return b
}

// Search adds new filters as search
func (b *Builder) Search(ai ...interface{}) *Builder {
	var f filter.Filter

	switch len(ai) {
	case 1:
		switch ai[0].(type) {
		case *filter.Filter:
			f = *ai[0].(*filter.Filter)
		default:
			f = *filter.Match(ai[0])
		}
	case 2:
		f = *filter.Match(ai[0].(string), ai[1])
	default:
		f = *filter.New(ai[0].(string), ai[1].(string), ai[2])
	}

	if b.BSearch == nil {
		b.BSearch = &[]filter.Filter{}
	}

	*b.BSearch = append(*b.BSearch, f)

	return b
}

// Sort by field and direction (asc, desc)
func (b *Builder) Sort(field string, direction ...string) *Builder {
	if b.BSort == nil {
		b.BSort = &[]map[string]string{}
	}

	dir := "asc"

	if direction != nil {
		dir = direction[0]
	}

	*b.BSort = append(*b.BSort, map[string]string{
		field: dir,
	})

	return b
}
