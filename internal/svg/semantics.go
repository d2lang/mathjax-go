// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import "github.com/d2lang/mathjax-go/internal/layout"

func (w *wrapper) computeSemanticsBBox(bbox *layout.BBox) {
	if len(w.children) == 0 {
		return
	}
	child := w.children[0].getBBox()
	bbox.W, bbox.H, bbox.D = child.W, child.H, child.D
}

func (w *wrapper) semanticsToSVG(parent *Element) {
	element := w.standardSVG(parent)
	if len(w.children) != 0 {
		w.children[0].toSVG(element)
	}
}

// SVGannotation has two intentional FIXMEs upstream: it produces its generic
// child SVG but leaves the bounding box at zero.  Semantics never selects an
// annotation for visual output, so this behavior is observable only when a
// source-shaped annotation is rendered directly.
func (w *wrapper) computeAnnotationBBox(bbox *layout.BBox) {
	bbox.W, bbox.H, bbox.D = 0, 0, 0
}

func (w *wrapper) annotationToSVG(parent *Element) {
	element := w.standardSVG(parent)
	w.addChildren(element)
}

func (w *wrapper) computeAnnotationXMLBBox(bbox *layout.BBox) {
	w.computeChildrenBBox(bbox)
}

func (w *wrapper) annotationXMLToSVG(parent *Element) {
	element := w.standardSVG(parent)
	x := 0.0
	for _, child := range w.children {
		if child.node.Kind == "XML" || child.node.Kind == "xml" {
			child.xmlToSVG(element)
		} else {
			child.toSVG(element)
		}
		bbox := child.outerBBox()
		child.place(x+bbox.L*bbox.RScale, 0, child.element)
		x += (bbox.L + bbox.W + bbox.R) * bbox.RScale
	}
}

// D2's LiteAdaptor reports an all-zero DOM bbox for embedded XML.  The shared
// Go MML node deliberately has no opaque browser-DOM payload, so the exact
// reachable compatibility behavior is a zero-size foreignObject container.
func (w *wrapper) computeXMLBBox(bbox *layout.BBox) {
	bbox.W, bbox.H, bbox.D = 0, 0, 0
}

func (w *wrapper) xmlToSVG(parent *Element) {
	bbox := w.getBBox()
	em := w.renderer.options.Em
	element := NewElement("foreignObject").
		SetAttr("data-mjx-xml", "true").
		SetAttr("y", jsFixed(-bbox.H*em, 1)+"px").
		SetAttr("width", jsFixed(bbox.W*em, 1)+"px").
		SetAttr("height", jsFixed((bbox.H+bbox.D)*em, 1)+"px").
		SetAttr("transform", "scale("+fixed(1/em)+") matrix(1 0 0 -1 0 0)")
	w.element = element
	parent.Append(element)
}
