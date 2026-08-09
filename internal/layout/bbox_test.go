// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"math"
	"testing"
)

func closeTo(got, want float64) bool { return math.Abs(got-want) < 1e-14 }

func TestBBoxCombineAppendAndUpdate(t *testing.T) {
	b := EmptyBBox()
	child := NewBBox(2, .7, .2)
	child.L, child.R, child.RScale = .1, .2, .5
	b.Combine(child, 1, .3)
	if !closeTo(b.W, 2.15) || !closeTo(b.H, .65) || !closeTo(b.D, -.2) {
		t.Fatalf("Combine() = %#v", b)
	}

	b = EmptyBBox()
	b.Append(child)
	if !closeTo(b.W, 1.15) || !closeTo(b.H, .35) || !closeTo(b.D, .1) {
		t.Fatalf("Append() = %#v", b)
	}

	other := NewBBox(4, 3, 2)
	other.PWidth = FullWidth
	b.UpdateFrom(other)
	if b.W != 4 || b.H != 3 || b.D != 2 || b.PWidth != FullWidth {
		t.Fatalf("UpdateFrom() = %#v", b)
	}
}

func TestBBoxCleanAndRescale(t *testing.T) {
	b := EmptyBBox()
	b.Clean()
	b.Rescale(2)
	if b.W != 0 || b.H != 0 || b.D != 0 || b.Scale != 1 || b.RScale != 1 {
		t.Fatalf("cleaned/rescaled box = %#v", b)
	}
}
