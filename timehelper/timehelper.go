package timehelper

import "time"

// RoundDuration rounds the duration to the nearest interval value
// from https://play.golang.org/p/QHocTHl8iR
// discussion: https://groups.google.com/forum/#!topic/golang-nuts/OWHmTBu16nA
// Original version by Andy Bursavich https://github.com/abursavich
func RoundDuration(duration, round time.Duration) time.Duration {
	if round <= 0 {
		return duration
	}

	negative := duration < 0

	if negative {
		duration *= -1
	}

	mid := duration % round
	duration -= mid

	if round <= 2*mid {
		duration += round
	}

	if negative {
		duration *= -1
	}

	return duration
}
