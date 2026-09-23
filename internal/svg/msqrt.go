// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func (w *wrapper) rootBaseIndex() int { return 0 }

func (w *wrapper) rootIndex() int {
	if w.node.Kind == "mroot" {
		return 1
	}
	return -1
}

func (w *wrapper) surdIndex() int {
	if w.node.Kind == "mroot" {
		return 2
	}
	return 1
}

func (w *wrapper) initializeRoot() {
	if len(w.children) == 0 {
		return
	}
	text := mml.NewText("√")
	mo := mml.NewNode("mo", nil, nil, text)
	mo.Flags.Token = true
	mo.Flags.Embellished = true
	mo.TeXClass = mml.TeXClassOrd
	mo.Attributes.Set("stretchy", true)
	surd := w.renderer.wrap(mo, w, w.scriptLevel, w.displayStyle)
	w.children = append(w.children, surd)
	surd.canStretch(font.DirectionVertical)
	base := w.children[w.rootBaseIndex()].outerBBox()
	t := w.renderer.params.Rule
	p := t
	if w.displayStyle {
		p = w.renderer.params.XHeight
	}
	w.surdHeight = base.H + base.D + 2*t + p/4
	surd.getStretchedVariant([]float64{w.surdHeight - base.D, base.D}, true)
}

func (w *wrapper) rootPQ(surd *layout.BBox) (p, q float64) {
	t := w.renderer.params.Rule
	p = t
	if w.displayStyle {
		p = w.renderer.params.XHeight
	}
	if surd.H+surd.D > w.surdHeight {
		q = ((surd.H + surd.D) - (w.surdHeight - 2*t - p/2)) / 2
	} else {
		q = t + p/4
	}
	return
}

func (w *wrapper) rootDimensions(surd *layout.BBox, height float64) (x, h, dx float64) {
	if w.node.Kind != "mroot" || w.rootIndex() >= len(w.children) {
		return 0, 0, 0
	}
	root := w.children[w.rootIndex()].outerBBox()
	surdWrapper := w.children[w.surdIndex()]
	offsetFactor := .6
	if surdWrapper.size < 0 {
		offsetFactor = .5
	}
	// JavaScript rounds both products before the final subtraction.
	offset := float64(offsetFactor * surd.W)
	width := math.Max(root.W, offset/root.RScale)
	dx = math.Max(0, width-root.W)
	total := surd.H + surd.D
	b := .55 * total
	if surdWrapper.size < 0 {
		b = 1.9
	}
	h = b - (total - height) + math.Max(0, root.D*root.RScale)
	x = float64(width*root.RScale) - offset
	return
}

func (w *wrapper) computeRootBBox(bbox *layout.BBox) {
	baseIndex, surdIndex := w.rootBaseIndex(), w.surdIndex()
	if baseIndex >= len(w.children) || surdIndex >= len(w.children) {
		w.computeChildrenBBox(bbox)
		return
	}
	surd := w.children[surdIndex].getBBox()
	base := w.children[baseIndex].outerBBox().Clone()
	_, q := w.rootPQ(surd)
	t := w.renderer.params.Rule
	height := base.H + q + t
	x, rootHeight, _ := w.rootDimensions(surd, height)
	bbox.H = height + t
	if rootIndex := w.rootIndex(); rootIndex >= 0 && rootIndex < len(w.children) {
		bbox.Combine(w.children[rootIndex].outerBBox(), 0, rootHeight)
	}
	bbox.Combine(surd, x, height-surd.H)
	bbox.Combine(base, x+surd.W, 0)
	bbox.Clean()
}

func (w *wrapper) rootToSVG(parent *Element) {
	baseIndex, surdIndex := w.rootBaseIndex(), w.surdIndex()
	if baseIndex >= len(w.children) || surdIndex >= len(w.children) {
		return
	}
	surd := w.children[surdIndex]
	base := w.children[baseIndex]
	surdBox := surd.getBBox()
	baseBox := base.outerBBox()
	_, q := w.rootPQ(surdBox)
	t := w.renderer.params.Rule * w.bbox.Scale
	height := baseBox.H + q + t
	element := w.standardSVG(parent)
	baseContainer := NewElement("g")
	element.Append(baseContainer)
	if rootIndex := w.rootIndex(); rootIndex >= 0 && rootIndex < len(w.children) {
		root := w.children[rootIndex]
		root.toSVG(element)
		x, rootHeight, dx := w.rootDimensions(surdBox, height)
		rootBox := root.outerBBox()
		root.place(dx*rootBox.RScale, rootHeight)
		w.dx = x
	}
	surd.toSVG(element)
	surd.place(w.dx, height-surdBox.H)
	base.toSVG(baseContainer)
	base.place(w.dx+surdBox.W, 0)
	element.Append(NewElement("rect").
		SetAttr("width", fixed(baseBox.W)).
		SetAttr("height", fixed(t)).
		SetAttr("x", fixed(w.dx+surdBox.W)).
		SetAttr("y", fixed(height-t)))
}
