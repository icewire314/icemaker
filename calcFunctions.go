package main

import (
	"math"
)

// add new funtions in this file
// also need to update map tables in main.go

func funcTwo(a, b float64, f func(float64, float64) float64) float64 {
	return f(a, b)
}

func funcOne(a float64, f func(float64) float64) float64 {
	return f(a)
}

func add(a, b float64) float64 {
	return a + b
}

func sub(a, b float64) float64 {
	return a - b
}

func mult(a, b float64) float64 {
	return a * b
}

func div(a, b float64) float64 {
	return a / b
}

func pow(a, b float64) float64 {
	return math.Pow(a, b)
}

func neg(a float64) float64 {
	return -1.0 * a
}

func pos(a float64) float64 {
	return a
}

func dBV(a float64) float64 {
	return 20 * math.Log10(a)
}

func dB(a float64) float64 {
	return 10 * math.Log10(a)
}

func atand(a float64) float64 { // returns arctan(a) in degrees
	// pi := 4 * math.Atan(1.0)
	return (180 / math.Pi) * math.Atan(a)
}

func parll(a, b float64) float64 { // parallel function for calculation of parallel resistors or series capacitors
	return 1 / (1/a + 1/b)
}
