// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package jscompat

import (
	"math"
	"testing"
)

func TestRound(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{2.5, 3},
		{-2.5, -2},
		{2.49, 2},
		{-2.51, -3},
	}
	for _, test := range tests {
		if got := Round(test.input); got != test.want {
			t.Errorf("Round(%v) = %v, want %v", test.input, got, test.want)
		}
	}
	negativeZero := Round(-0.25)
	if negativeZero != 0 || !math.Signbit(negativeZero) {
		t.Errorf("Round(-0.25) = %v, want negative zero", negativeZero)
	}
}

func TestToFixed(t *testing.T) {
	tests := []struct {
		input  float64
		digits int
		want   string
	}{
		{2.5, 0, "3"},
		{-2.5, 0, "-3"},
		{1.25, 1, "1.3"},
		{-1.25, 1, "-1.3"},
		{1.005, 2, "1.00"},
		{0, 3, "0.000"},
		{math.Copysign(0, -1), 3, "0.000"},
		{1e21, 3, "1e+21"},
	}
	for _, test := range tests {
		if got := ToFixed(test.input, test.digits); got != test.want {
			t.Errorf("ToFixed(%v, %d) = %q, want %q", test.input, test.digits, got, test.want)
		}
	}
}

func TestNumberString(t *testing.T) {
	tests := map[float64]string{
		0:       "0",
		1e-7:    "1e-7",
		1e-6:    "0.000001",
		1e20:    "100000000000000000000",
		1e21:    "1e+21",
		-1.5e-7: "-1.5e-7",
	}
	for input, want := range tests {
		if got := NumberString(input); got != want {
			t.Errorf("NumberString(%v) = %q, want %q", input, got, want)
		}
	}
}
