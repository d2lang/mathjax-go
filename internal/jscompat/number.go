// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 behavior.

// Package jscompat implements JavaScript numeric behavior that is observable
// in MathJax 3.2.2's SVG output.
package jscompat

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// NumberStep forces an intermediate float64 rounding point. It prevents the Go
// compiler from fusing operations that JavaScript evaluates separately.
//
//go:noinline
func NumberStep(value float64) float64 { return value }

// Round implements ECMAScript Math.round, including negative zero for values
// in [-0.5, 0).
func Round(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value == 0 || math.Abs(value) >= 1<<52 {
		return value
	}
	if value >= -0.5 && value < 0 {
		return math.Copysign(0, -1)
	}
	return math.Floor(value + 0.5)
}

// TruthyNumber reports JavaScript truthiness for a Number.
func TruthyNumber(value float64) bool { return value != 0 && !math.IsNaN(value) }

// NumberString formats a float64 as ECMAScript String(Number) for the finite
// and non-finite values observable in MathJax's serialization paths.
func NumberString(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	}

	abs := math.Abs(value)
	if abs >= 1e-6 && abs < 1e21 {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	text := strconv.FormatFloat(value, 'e', -1, 64)
	mantissa, exponent, ok := strings.Cut(text, "e")
	if !ok {
		return text
	}
	sign := ""
	if strings.HasPrefix(exponent, "+") || strings.HasPrefix(exponent, "-") {
		sign, exponent = exponent[:1], exponent[1:]
	}
	exponent = strings.TrimLeft(exponent, "0")
	if exponent == "" {
		exponent = "0"
	}
	return mantissa + "e" + sign + exponent
}

// ToFixed implements Number.prototype.toFixed for 0 through 100 fraction
// digits. Unlike strconv's fixed formatting, ECMAScript resolves an exact tie
// toward the larger integer after operating on the Number's exact binary
// value. That difference is visible for values such as 2.5.
func ToFixed(value float64, digits int) string {
	if digits < 0 || digits > 100 {
		return "RangeError"
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return NumberString(value)
	}
	if math.Abs(value) >= 1e21 {
		return NumberString(value)
	}

	negative := value < 0
	if negative {
		value = -value
	}

	rational := new(big.Rat).SetFloat64(value)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	numerator := new(big.Int).Mul(rational.Num(), scale)
	denominator := rational.Denom()
	integer, remainder := new(big.Int), new(big.Int)
	integer.QuoRem(numerator, denominator, remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(denominator) >= 0 {
		integer.Add(integer, big.NewInt(1))
	}

	text := integer.String()
	if digits > 0 {
		if len(text) <= digits {
			text = strings.Repeat("0", digits-len(text)+1) + text
		}
		point := len(text) - digits
		text = text[:point] + "." + text[point:]
	}
	if negative {
		text = "-" + text
	}
	return text
}

// Fixed reproduces CommonOutputJax.fixed(): values close to zero collapse to
// "0", and trailing fractional zeroes are removed.
func Fixed(value float64, digits int) string {
	if math.Abs(value) < 0.0006 {
		return "0"
	}
	return trimFixed(ToFixed(value, digits))
}

// TrimFixed removes the trailing zeroes and optional decimal point in a
// toFixed result, as MathJax does with /\.?0+$/. It is exported for the SVG
// output package's other fixed-precision helpers.
func TrimFixed(text string) string { return trimFixed(text) }

func trimFixed(text string) string {
	if !strings.Contains(text, ".") {
		return text
	}
	text = strings.TrimRight(text, "0")
	return strings.TrimSuffix(text, ".")
}
