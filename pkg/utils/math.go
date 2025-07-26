package utils

import "math"

const epsilon = 1e-9

func IsZero(f float64) bool {
	return math.Abs(f) < epsilon
}

func Round(f float64, n int) float64 {
	power := math.Pow10(n)
	return math.Round(f*power) / power
}
