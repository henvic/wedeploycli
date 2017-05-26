package timehelper

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

var compare = `     duration            ms          0.5s             s           10s             m             h
       9.63ms          10ms            0s            0s            0s            0s            0s
  1.23456789s        1.235s            1s            1s            0s            0s            0s
         1.5s          1.5s          1.5s            2s            0s            0s            0s
         1.4s          1.4s          1.5s            1s            0s            0s            0s
        -1.4s         -1.4s         -1.5s           -1s            0s            0s            0s
        -1.5s         -1.5s         -1.5s           -2s            0s            0s            0s
     8.91234s        8.912s            9s            9s           10s            0s            0s
    34.56789s       34.568s         34.5s           35s           30s          1m0s            0s
3h25m45.6789s  3h25m45.679s    3h25m45.5s      3h25m46s      3h25m50s       3h26m0s        3h0m0s
`

func TestRoundDuration(t *testing.T) {
	got := RoundDuration(926*time.Second, time.Minute)
	want := 15 * time.Minute
	wantText := "15m0s"

	if want != got {
		t.Errorf("Expected duration to be %v, got %v instead", want, got)
	}

	if wantText != got.String() {
		t.Errorf("Expected duration text to be %v, got %v instead", want, got)
	}
}
func TestRoundDurationNegative(t *testing.T) {
	got := RoundDuration(time.Second, -10*time.Hour)
	want := time.Second
	wantText := "1s"

	if want != got {
		t.Errorf("Expected duration to be %v, got %v instead", want, got)
	}

	if wantText != got.String() {
		t.Errorf("Expected duration text to be %v, got %v instead", want, got)
	}
}

func TestRoundDurationText(t *testing.T) {
	var b bytes.Buffer
	samples := []time.Duration{9.63e6, 1.23456789e9, 1.5e9, 1.4e9, -1.4e9, -1.5e9, 8.91234e9, 34.56789e9, 12345.6789e9}
	format := "% 13s % 13s % 13s % 13s % 13s % 13s % 13s\n"
	fmt.Fprintf(&b, format, "duration", "ms", "0.5s", "s", "10s", "m", "h")
	for _, d := range samples {
		fmt.Fprintf(
			&b,
			format,
			d,
			RoundDuration(d, time.Millisecond),
			RoundDuration(d, 0.5e9),
			RoundDuration(d, time.Second),
			RoundDuration(d, 10*time.Second),
			RoundDuration(d, time.Minute),
			RoundDuration(d, time.Hour),
		)
	}

	if b.String() != compare {
		t.Errorf("Expected duration to be equal to template")
	}
}
