// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/layout"
)

// mathAlignShift is CommonWrapper.getAlignShift for the math wrapper only.
// The output options currently expose MathJax's default center/zero alignment.
// Keep existing table layout consumers independent of this late root policy.
func (w *wrapper) mathAlignShift() (string, float64) {
	align := stringAttribute(w.node, "indentalign", "auto")
	first := stringAttribute(w.node, "indentalignfirst", "indentalign")
	if first != "indentalign" {
		align = first
	}
	if align == "auto" {
		align = "center"
	}
	shift := attribute(w.node, "indentshift", "auto")
	firstShift := attribute(w.node, "indentshiftfirst", "indentshift")
	if firstShift != "indentshift" {
		shift = firstShift
	}
	if shift == "auto" {
		// The fixed displayIndent "0" satisfies the primary zero-unit guard,
		// including right alignment. Explicit shifts are never sign-reversed.
		shift = "0"
	}
	// getAlignShift passes raw containerWidth, unlike getWrapWidth's em size.
	// The pinned conversion's default container is 80ex in CSS pixels.
	return align, layout.Length2Em(shift, 80*w.renderer.options.Ex, w.bbox.Scale, w.renderer.pxPerEm)
}

// handleMathMinWidth follows SVGmath.handleDisplay's responsive-width branch.
func (w *wrapper) handleMathMinWidth() {
	if w.bbox.PWidth != layout.FullWidth || w.renderer.table == nil {
		return
	}
	align, shift := w.mathAlignShift()
	box := w.renderer.table.outerBBox()
	left, width, right := box.L, box.W, box.R
	switch align {
	case "right":
		if right == 0 || math.IsNaN(right) {
			right = -shift
		}
		right = math.Max(right, -shift)
	case "left":
		if left == 0 || math.IsNaN(left) {
			left = shift
		}
		left = math.Max(left, shift)
	case "center":
		width += 2 * math.Abs(shift)
	}
	w.renderer.minWidth = math.Max(0, left+width+right)
}
