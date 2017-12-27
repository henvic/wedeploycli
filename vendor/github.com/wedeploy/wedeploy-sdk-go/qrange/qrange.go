package qrange

// Range type
type Range struct {
	From *int `json:"from,omitempty"`
	To   *int `json:"to,omitempty"`
}

// Between creates a range between an interval
func Between(from, to int) Range {
	return new(&from, &to)
}

// From creates a range from a given time until present
func From(from int) Range {
	return new(&from, nil)
}

// To creates a range from the beginning of the universe until the given time
func To(to int) Range {
	return new(nil, &to)
}

func new(from, to *int) Range {
	return Range{
		From: from,
		To:   to,
	}
}
