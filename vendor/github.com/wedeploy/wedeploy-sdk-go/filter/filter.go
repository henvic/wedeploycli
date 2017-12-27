package filter

import (
	"github.com/wedeploy/api-go/geo"
	"github.com/wedeploy/api-go/qrange"
)

// Filter is a map with the filter data
type Filter map[string]interface{}

type data struct {
	Operator string      `json:"operator"`
	Value    interface{} `json:"value,omitempty"`
}

// New creates a new New filter
func New(field, operator string, value interface{}) *Filter {
	var m = make(Filter)

	m[field] = &data{
		Operator: operator,
		Value:    value,
	}

	return &m
}

// Equal creates a new Equal filter
func Equal(field string, value interface{}) *Filter {
	return New(field, "=", value)
}

// NotEqual creates a new NotEqual filter
func NotEqual(field string, value interface{}) *Filter {
	return New(field, "!=", value)
}

// Gt creates a new Gt filter
func Gt(field string, value interface{}) *Filter {
	return New(field, ">", value)
}

// Gte creates a new Gte filter
func Gte(field string, value interface{}) *Filter {
	return New(field, ">=", value)
}

// Lt creates a new Lt filter
func Lt(field string, value interface{}) *Filter {
	return New(field, "<", value)
}

// Lte creates a new Lte filter
func Lte(field string, value interface{}) *Filter {
	return New(field, "=<", value)
}

// Regex creates a new Regex filter
func Regex(field string, value interface{}) *Filter {
	return New(field, "~", value)
}

// None creates a new None filter
func None(field string, value interface{}) *Filter {
	return New(field, "none", value)
}

// Any creates a new Any filter
func Any(field string, value ...interface{}) *Filter {
	return New(field, "any", value)
}

// Add creates a new Add filter
func Add(operator string, filter ...*Filter) *Filter {
	m := make(Filter)
	m[operator] = filter
	return &m
}

// Add creates and return a new Add filter from the existing filter
func (f *Filter) Add(args ...interface{}) *Filter {
	m := make(Filter)

	var operator = args[0].(string)

	switch len(args) {
	case 1:
		m[operator] = f
	case 4:
		m = *Add(operator, f, New(args[1].(string), args[2].(string), args[3]))
	default:
		args = append(args[:0], args[1:]...)
		var filters = []*Filter{f}

		for _, filter := range args {
			filters = append(filters, filter.(*Filter))
		}

		m = *Add(operator, filters...)
	}

	return &m
}

// And creates a new And filter
func And(filter ...*Filter) *Filter {
	return Add("and", filter...)
}

// Or creates a new Or filter
func Or(filter ...*Filter) *Filter {
	return Add("or", filter...)
}

// Exists creates a new Exists filter
func Exists(field string) *Filter {
	return New(field, "exists", nil)
}

// Missing creates a new Missing filter
func Missing(field string) *Filter {
	return New(field, "missing", nil)
}

// Match creates a new Match filter
func Match(args ...interface{}) *Filter {
	var field string
	var value interface{}
	switch len(args) {
	case 1:
		field = "*"
		value = args[0]
	case 2:
		field = args[0].(string)
		value = args[1]
	}
	return New(field, "match", value)
}

// Phrase creates a new Phrase filter
func Phrase(value string) *Filter {
	return New("*", "phrase", value)
}

// Prefix creates a new Prefix filter
func Prefix(value string) *Filter {
	return New("*", "prefix", value)
}

// Fuzzy creates a new Fuzzy filter
func Fuzzy(
	fieldOrQuery string, query interface{}, fuzziness interface{}) *Filter {
	return q("fuzzy", fieldOrQuery, query, fuzziness)
}

// Similar creates a new Similar filter
func Similar(
	fieldOrQuery string, query interface{}, fuzziness interface{}) *Filter {
	return q("similar", fieldOrQuery, query, fuzziness)
}

// Distance creates a new Distance filter
func Distance(field string, location interface{}, lr interface{}) *Filter {
	value := make(map[string]interface{})

	switch location.(type) {
	case geo.Circle:
		geoCircles := location.(geo.Circle)
		value["location"] = geoCircles.Coordinates
		value["max"] = geoCircles.Radius
	default:
		value["location"] = location.(geo.Point)

		switch lr.(type) {
		case qrange.Range:
			r := lr.(qrange.Range)

			if r.From != nil {
				value["min"] = r.From
			}

			if r.To != nil {
				value["max"] = r.To
			}
		case int:
			value["max"] = lr.(int)
		}

	}

	return New(field, "gd", value)
}

// Range creates a new Range filter
func Range(field string, args ...interface{}) *Filter {
	if len(args) == 1 {
		return New(field, "range", args[0].(qrange.Range))
	}

	return New(field, "range", qrange.Between(args[0].(int), args[1].(int)))
}

// Shape creates a new Shape filter
func Shape(field string, shapes ...interface{}) *Filter {
	value := make(map[string]interface{})
	value["type"] = "geometrycollection"
	value["geometries"] = shapes
	return New(field, "gs", value)
}

// Polygon creates a new Polygon filter
func Polygon(field string, points ...geo.Point) *Filter {
	return New(field, "gp", points)
}

// BoundingBox creates a new BoundingBox filter
func BoundingBox(
	field string, boxOrUpperLeft interface{}, lowerRight ...interface{}) *Filter {
	var coords []geo.Point
	switch boxOrUpperLeft.(type) { // or len(lowerRight)
	case geo.BoundingBox: // or 0
		coords = boxOrUpperLeft.(geo.BoundingBox).Coordinates
	default:
		coords = []geo.Point{
			boxOrUpperLeft.(geo.Point),
			lowerRight[0].(geo.Point),
		}
	}

	return Polygon(field, coords...)
}

func q(
	qType, fieldOrQuery string, query interface{}, fuzziness interface{}) *Filter {
	var field string

	switch query {
	case nil:
		field = "*"
		query = fieldOrQuery
	default:
		field = fieldOrQuery
	}

	value := make(map[string]interface{})

	value["query"] = query

	if fuzziness != nil {
		value["fuzziness"] = fuzziness
	}

	return New(field, qType, value)
}
