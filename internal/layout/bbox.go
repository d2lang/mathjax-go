// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

// Package layout implements MathJax's common TeX layout rules.
package layout

// BigDimen is MathJax's sentinel for an unset or effectively infinite
// dimension.
const BigDimen = 1_000_000

// FullWidth is the percentage-width marker used by MathJax tables.
const FullWidth = "100%"

// BBox is MathJax's common-output bounding box. Dimensions are in ems.
// L and R are space outside W, RScale is relative to the parent, IC is the
// italic correction, Skew is the accent skew, and DX is the combining-accent
// offset.
type BBox struct {
	W, H, D float64

	Scale  float64
	RScale float64
	L, R   float64
	PWidth string
	IC     float64
	Skew   float64
	DX     float64
}

// NewBBox returns the initialized zero-width box used by MathJax. An omitted
// height or depth is represented by -BigDimen.
func NewBBox(w, h, d float64) *BBox {
	return &BBox{W: w, H: h, D: d, Scale: 1, RScale: 1}
}

// ZeroBBox returns a fully specified zero box.
func ZeroBBox() *BBox { return NewBBox(0, 0, 0) }

// EmptyBBox returns a box whose height and depth have not been set.
func EmptyBBox() *BBox { return NewBBox(0, -BigDimen, -BigDimen) }

// Empty prepares b for Append and Combine operations.
func (b *BBox) Empty() *BBox {
	b.W = 0
	b.H = -BigDimen
	b.D = -BigDimen
	return b
}

// Clean replaces unset dimensions with zero.
func (b *BBox) Clean() {
	if b.W == -BigDimen {
		b.W = 0
	}
	if b.H == -BigDimen {
		b.H = 0
	}
	if b.D == -BigDimen {
		b.D = 0
	}
}

// Rescale multiplies the three geometric dimensions. It intentionally does
// not modify spaces or scale metadata, matching MathJax 3.2.2.
func (b *BBox) Rescale(scale float64) {
	b.W *= scale
	b.H *= scale
	b.D *= scale
}

// Combine expands b to include child at offset (x,y).
func (b *BBox) Combine(child *BBox, x, y float64) {
	scale := child.RScale
	w := x + scale*(child.W+child.L+child.R)
	h := y + scale*child.H
	d := scale*child.D - y
	if w > b.W {
		b.W = w
	}
	if h > b.H {
		b.H = h
	}
	if d > b.D {
		b.D = d
	}
}

// Append places child after the current width and expands height and depth.
func (b *BBox) Append(child *BBox) {
	scale := child.RScale
	b.W += scale * (child.W + child.L + child.R)
	if scale*child.H > b.H {
		b.H = scale * child.H
	}
	if scale*child.D > b.D {
		b.D = scale * child.D
	}
}

// UpdateFrom overwrites b's geometric dimensions and copies a non-empty
// percentage width.
func (b *BBox) UpdateFrom(other *BBox) {
	b.H = other.H
	b.D = other.D
	b.W = other.W
	if other.PWidth != "" {
		b.PWidth = other.PWidth
	}
}

// Clone returns an independent copy.
func (b *BBox) Clone() *BBox {
	if b == nil {
		return nil
	}
	clone := *b
	return &clone
}
