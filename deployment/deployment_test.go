package deployment

import (
	"testing"
)

func TestUpdateMessageErrorStringCounter(t *testing.T) {
	var msgs = map[string]string{
		"":                       "(error #1)",
		"oi":                     "oi (error #1)",
		"oi (error #1)":          "oi (error #2)",
		"oi (error #2)":          "oi (error #3)",
		"oi (error #3)":          "oi (error #4)",
		"(error #1)":             "(error #2)",
		"(error #1) (error #1)":  "(error #2) (error #2)",
		"(error #6) (error #3)":  "(error #7) (error #4)",
		"(error #20)":            "(error #21)",
		"(error #20) xyz":        "(error #21) xyz",
		"abc (error #20) xyz":    "abc (error #21) xyz",
		"abc (error #21) xyz":    "abc (error #22) xyz",
		"abc (error #12321) xyz": "abc (error #12322) xyz",
	}

	for k, v := range msgs {
		if got := updateMessageErrorStringCounter(k); got != v {
			t.Errorf("Expected message to be %v, got %v instead", v, got)
		}
	}
}
