package main

import (
	"math"
	"time"
)

func float64ToTime(f float64) time.Time {
	sec, dec := math.Modf(f)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}
