// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/output/{common,svg}/Wrappers/mfenced.ts.

package svg

import (
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

// CommonMfenced creates an empty inferred MML row, then interleaves wrappers
// for its already-inherited fake operators and authored children. Their MML
// parents remain mfenced; a semantic mrow would change MathML spacing.
func (w *wrapper) initializeFencedRow() {
	row := syntheticMMLFactory.Create("inferredMrow")
	row.Kind = "mrow"
	row.Attributes.SetInherited("displaystyle", attribute(w.node, "displaystyle", w.displayStyle))
	row.Attributes.SetInherited("scriptlevel", attribute(w.node, "scriptlevel", w.scriptLevel))
	if w.node.Attributes.IsSet("mathsize") {
		row.Attributes.SetInherited("mathsize", attribute(w.node, "mathsize", "normal"))
	}
	if prime, ok := w.node.Property("texprimestyle"); ok && truthy(prime) {
		row.SetProperty("texprimestyle", true)
	}
	w.fencedRow = w.renderer.wrap(row, w, w.scriptLevel, w.displayStyle)
	addMo := func(node *mml.Node) {
		if node == nil {
			return
		}
		child := w.renderer.wrap(node, w, w.scriptLevel, w.displayStyle)
		w.fencedRow.children = append(w.fencedRow.children, child)
		child.parent = w.fencedRow
	}
	if w.node.Fenced != nil {
		addMo(w.node.Fenced.Open)
	}
	for i, child := range w.children {
		if i > 0 && w.node.Fenced != nil && i-1 < len(w.node.Fenced.Separators) {
			addMo(w.node.Fenced.Separators[i-1])
		}
		w.fencedRow.children = append(w.fencedRow.children, child)
	}
	if w.node.Fenced != nil {
		addMo(w.node.Fenced.Close)
	}
	w.fencedRow.stretchRowChildren()
}

func (w *wrapper) computeFencedBBox(bbox *layout.BBox) {
	bbox.UpdateFrom(w.fencedRow.outerBBox())
}

func (w *wrapper) fencedToSVG(parent *Element) {
	element := w.standardSVG(parent)
	// SVGmfenced changes only wrapper ownership during output. Neither its
	// real MML children nor the fake operators join the temporary row's node.
	for _, child := range w.children {
		child.parent = w.fencedRow
	}
	w.fencedRow.toSVG(element)
	for _, child := range w.children {
		child.parent = w
	}
}
