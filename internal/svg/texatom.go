// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func (w *wrapper) computeTeXAtomBBox(bbox *layout.BBox) {
	w.computeChildrenBBox(bbox)
	if len(w.children) != 0 {
		if correction := w.children[0].bbox.IC; correction != 0 {
			bbox.IC = correction
		}
	}
	if effectiveTeXClass(w.node) == mml.TeXClassVCenter {
		delta := (bbox.H+bbox.D)/2 + w.renderer.params.Axis - bbox.H
		bbox.H += delta
		bbox.D -= delta
	}
}

func (w *wrapper) texAtomToSVG(parent *Element) {
	element := w.standardSVG(parent)
	w.addChildren(element)
	class := effectiveTeXClass(w.node)
	if class >= mml.TeXClassOrd && int(class) < len(mml.TeXClassNames) {
		element.SetAttr("data-mjx-texclass", mml.TeXClassNames[class])
	}
	if class == mml.TeXClassVCenter && len(w.children) != 0 {
		bbox := w.children[0].getBBox()
		delta := (bbox.H+bbox.D)/2 + w.renderer.params.Axis - bbox.H
		element.SetAttr("transform", "translate(0 "+fixed(delta)+")")
	}
}
