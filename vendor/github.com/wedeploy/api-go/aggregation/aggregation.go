package aggregation

import "github.com/wedeploy/api-go/qrange"

// Aggregation is a map with the aggregation data
type Aggregation map[string]*data

type data struct {
	Name     string      `json:"name"`
	Operator interface{} `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// Avg creates and return a new aggregation
func Avg(name, field string) *Aggregation {
	return New(name, field, "avg", nil)
}

// Count creates and return a new count aggregation
func Count(name, field string) *Aggregation {
	return New(name, field, "count", nil)
}

// Distance creates and return a new distance aggregation
func Distance(
	name, field string,
	location interface{},
	lr ...qrange.Range) *Aggregation {
	value := make(map[string]interface{})

	value["location"] = location
	value["ranges"] = lr

	return New(name, field, "geoDistance", value)
}

// ExtendedStats creates and return a new extendedStats aggregation
func ExtendedStats(name, field string) *Aggregation {
	return New(name, field, "extendedStats", nil)
}

// Histogram creates and return a new histogram aggregation
func Histogram(name, field string, interval int) *Aggregation {
	return New(name, field, "histogram", interval)
}

// Max creates and return a new max aggregation
func Max(name, field string) *Aggregation {
	return New(name, field, "max", nil)
}

// Min creates and return a new min aggregation
func Min(name, field string) *Aggregation {
	return New(name, field, "min", nil)
}

// Missing creates and return a new missing aggregation
func Missing(name, field string) *Aggregation {
	return New(name, field, "missing", nil)
}

// New creates and return a new aggregation
func New(
	name, field string,
	operator interface{},
	value interface{}) *Aggregation {
	var m = make(Aggregation)

	m[field] = &data{
		Name:     name,
		Operator: operator,
		Value:    value,
	}

	return &m
}

// Stats creates and return a new stats aggregation
func Stats(name, field string) *Aggregation {
	return New(name, field, "stats", nil)
}

// Sum creates and return a new sum aggregation
func Sum(name, field string) *Aggregation {
	return New(name, field, "sum", nil)
}

// Terms creates and return a new terms aggregation
func Terms(name, field string) *Aggregation {
	return New(name, field, "terms", nil)
}

// Range sets a range for the aggregation data
func (a *Aggregation) Range(args ...interface{}) *Aggregation {
	var i = (*a)[a.getFieldName()].Value
	var ra qrange.Range

	switch len(args) {
	case 2:
		ra = qrange.Between(args[0].(int), args[1].(int))
	default:
		ra = args[0].(qrange.Range)
	}

	var x = i.(map[string]interface{})["ranges"]
	var r = x.([]qrange.Range)
	i.(map[string]interface{})["ranges"] = append(r, ra)

	return a
}

// Unit sets the unit for the aggregation data
func (a *Aggregation) Unit(unit string) *Aggregation {
	var i = (*a)[a.getFieldName()].Value
	i.(map[string]interface{})["unit"] = unit
	return a
}

func (a *Aggregation) getFieldName() string {
	var field string
	for k := range *a {
		field = k
	}
	return field
}
