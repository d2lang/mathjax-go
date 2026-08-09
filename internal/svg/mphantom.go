// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

// phantomToSVG intentionally renders only the standard wrapper group.  Its
// children still contribute to layout through the ordinary bounding-box path.
func (w *wrapper) phantomToSVG(parent *Element) {
	w.standardSVG(parent)
}
