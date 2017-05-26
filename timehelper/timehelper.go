package timehelper

import "time"

// RoundDuration rounds the duration to the nearest interval value
// from https://play.golang.org/p/QHocTHl8iR
func RoundDuration(d, r time.Duration) time.Duration {

	if r <= 0 {
		return d
	}

	neg := d < 0
	if neg {
		d = -d
	}

	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}

	if neg {
		return -d
	}

	return d
}
