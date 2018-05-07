package feedback

import (
	"testing"
)

func TestUpdateMessageErrorStringCounter(t *testing.T) {
	var msgs = map[string]string{
		"":   "(retrying to get status #1)",
		"oi": "oi (retrying to get status #1)",
		"oi (retrying to get status #1)":                          "oi (retrying to get status #2)",
		"oi (retrying to get status #2)":                          "oi (retrying to get status #3)",
		"oi (retrying to get status #3)":                          "oi (retrying to get status #4)",
		"(retrying to get status #1)":                             "(retrying to get status #2)",
		"(retrying to get status #1) (retrying to get status #1)": "(retrying to get status #2) (retrying to get status #2)",
		"(retrying to get status #6) (retrying to get status #3)": "(retrying to get status #7) (retrying to get status #4)",
		"(retrying to get status #20)":                            "(retrying to get status #21)",
		"(retrying to get status #20) xyz":                        "(retrying to get status #21) xyz",
		"abc (retrying to get status #20) xyz":                    "abc (retrying to get status #21) xyz",
		"abc (retrying to get status #21) xyz":                    "abc (retrying to get status #22) xyz",
		"abc (retrying to get status #12321) xyz":                 "abc (retrying to get status #12322) xyz",
	}

	for k, v := range msgs {
		if got := updateMessageErrorStringCounter(k); got != v {
			t.Errorf("Expected message to be %v, got %v instead", v, got)
		}
	}
}
