// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

// errorToSVG ports SVGmerror.toSVG.  In addition to its content, merror
// carries a yellow background rectangle and exposes the TeX parser's message
// through data-mjx-error.
func (w *wrapper) errorToSVG(parent *Element) {
	element := w.standardSVG(parent)
	if message := stringAttributeExplicit(w.node, "data-mjx-error"); message != "" {
		element.SetAttr("data-mjx-error", message)
	}
	bbox := w.getBBox()
	element.Append(NewElement("rect").
		SetAttr("data-background", "true").
		SetAttr("width", fixed(bbox.W)).
		SetAttr("height", fixed(bbox.H+bbox.D)).
		SetAttr("y", fixed(-bbox.D)))
	if title := stringAttribute(w.node, "title", ""); title != "" {
		element.Append(NewElement("title", Text(title)))
	}
	w.addChildren(element)
}
