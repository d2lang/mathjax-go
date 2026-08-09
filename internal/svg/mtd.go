// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import "github.com/d2lang/mathjax-go/internal/layout"

func (w *wrapper) computeCellBBox(bbox *layout.BBox) {
	w.computeChildrenBBox(bbox)
}

func (w *wrapper) cellToSVG(parent *Element) {
	element := w.standardSVG(parent)
	w.addChildren(element)
}

func (w *wrapper) resizeCellBackground(x, y, width, height float64) {
	background := tableBackground(w.element)
	if background == nil {
		return
	}
	background.SetAttr("x", fixed(x))
	background.SetAttr("y", fixed(y))
	background.SetAttr("width", fixed(width))
	background.SetAttr("height", fixed(height))
}
