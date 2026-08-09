// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import "fmt"

// linkParent ports SVGWrapper.createSVGnode's href path.  MathJax wraps the
// node in an SVG anchor and installs a transparent hit box as the first child
// of the linked group so the entire measured box remains clickable.
func (w *wrapper) linkParent(parent, element *Element) *Element {
	if w.node == nil || w.node.Attributes == nil {
		return parent
	}
	value, ok := w.node.Attributes.Get("href")
	if !ok || fmt.Sprint(value) == "" {
		return parent
	}
	anchor := NewElement("a").SetAttr("href", fmt.Sprint(value))
	parent.Append(anchor)
	bbox := w.outerBBox()
	element.Append(NewElement("rect").
		SetAttr("data-hitbox", "true").
		SetAttr("fill", "none").
		SetAttr("stroke", "none").
		SetAttr("pointer-events", "all").
		SetAttr("width", fixed(bbox.W)).
		SetAttr("height", fixed(bbox.H+bbox.D)).
		SetAttr("y", fixed(-bbox.D)))
	return anchor
}
