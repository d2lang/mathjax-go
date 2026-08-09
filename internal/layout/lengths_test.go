// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"math"
	"testing"
)

func TestLength2Em(t *testing.T) {
	tests := []struct {
		length any
		size   float64
		want   float64
	}{
		{"", 3, 3},
		{nil, 3, 3},
		{"thinmathspace", 0, 3.0 / 18},
		{"2em", 0, 2},
		{"3ex", 0, 1.293},
		{"50%", 8, 4},
		{"2", 3, 6},
		{"bogus", 4, 4},
		{" 1in trailing", 0, 6},
	}
	for _, test := range tests {
		got := Length2Em(test.length, test.size, 1, 16)
		if math.Abs(got-test.want) > 1e-12 {
			t.Errorf("Length2Em(%#v) = %v, want %v", test.length, got, test.want)
		}
	}
}

func TestLengthFormatting(t *testing.T) {
	if got := Percent(.125); got != "12.5%" {
		t.Errorf("Percent = %q", got)
	}
	if got := Em(.0009); got != "0" {
		t.Errorf("Em = %q", got)
	}
	if got := Em(-1.2345); got != "-1.234em" {
		t.Errorf("Em = %q", got)
	}
	if got := EmRounded(1.03, 16); got != "1.003em" {
		t.Errorf("EmRounded = %q", got)
	}
	if got := PX(.01, 1.25, 16); got != "1.3px" {
		t.Errorf("PX = %q", got)
	}
}
