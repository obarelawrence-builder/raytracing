package main

import "math"

// Interval is a closed range [Min, Max].
// Used instead of passing tMin/tMax as separate floats everywhere.
type Interval struct {
	Min, Max float64
}

// EmptyInterval has no valid values.
var EmptyInterval = Interval{math.Inf(1), math.Inf(-1)}

// UniverseInterval covers all real numbers.
var UniverseInterval = Interval{math.Inf(-1), math.Inf(1)}

// Contains reports whether x is within [Min, Max] (inclusive).
func (i Interval) Contains(x float64) bool { return i.Min <= x && x <= i.Max }

// Surrounds reports whether x is strictly inside (Min, Max).
func (i Interval) Surrounds(x float64) bool { return i.Min < x && x < i.Max }

// Clamp clamps x to the interval [Min, Max].
func (i Interval) Clamp(x float64) float64 {
	if x < i.Min {
		return i.Min
	}
	if x > i.Max {
		return i.Max
	}
	return x
}
