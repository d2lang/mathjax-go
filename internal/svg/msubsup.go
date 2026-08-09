// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/layout"
)

func (w *wrapper) scriptBase() *wrapper {
	if len(w.children) == 0 {
		return nil
	}
	return w.children[0]
}

func (w *wrapper) scriptBaseCore() *wrapper {
	core := w.scriptBase()
	for core != nil && len(core.children) == 1 {
		switch core.node.Kind {
		case "mrow", "mstyle", "mpadded", "mphantom", "semantics":
			core = core.children[0]
		case "TeXAtom":
			if effectiveTeXClass(core.node) != 8 {
				core = core.children[0]
				continue
			}
			return core
		default:
			return core
		}
	}
	return core
}

func (w *wrapper) baseScale() float64 {
	core := w.scriptBaseCore()
	scale := 1.0
	for current := core; current != nil && current != w; current = current.parent {
		scale *= current.outerBBox().RScale
	}
	return scale
}

func (w *wrapper) baseIC() float64 {
	core := w.scriptBaseCore()
	if core == nil {
		return 0
	}
	return core.outerBBox().IC * w.baseScale()
}

func (w *wrapper) baseIsChar() bool {
	core := w.scriptBaseCore()
	if core == nil || (core.node.Kind != "mi" && core.node.Kind != "mn" && core.node.Kind != "mo") {
		return false
	}
	if core.node.Kind == "mo" {
		// CommonScriptbase.isCharBase only treats an unstretched mo (size ===
		// null) as a character.  A fixed-size delimiter such as \bigr has a
		// selected size and must retain the TeX sup-drop term.
		if core.sizeSet || boolAttributeDefault(core.node, "largeop", false) {
			return false
		}
	}
	return core.bbox.RScale == 1 && utf8.RuneCountInString(nodeText(core.node)) == 1
}

func (w *wrapper) removeBaseIC() bool {
	return w.node.Kind == "msub" || w.node.Kind == "msubsup" ||
		w.node.Kind == "munder" || w.node.Kind == "munderover"
}

func (w *wrapper) baseWidth() float64 {
	base := w.scriptBase()
	if base == nil {
		return 0
	}
	bbox := base.outerBBox()
	width := bbox.W * bbox.RScale
	if w.removeBaseIC() {
		width -= w.baseIC()
	}
	return width + w.renderer.params.ExtraIC
}

func (w *wrapper) adjustedIC() float64 {
	core := w.scriptBaseCore()
	if core == nil {
		return 0
	}
	ic := core.outerBBox().IC
	if ic == 0 {
		return 0
	}
	return (1.05*ic + .05) * w.baseScale()
}

func (w *wrapper) baseCharZero(value float64) float64 {
	core := w.scriptBaseCore()
	largeop := core != nil && boolAttributeDefault(core.node, "largeop", false)
	if w.baseIsChar() && !largeop && w.baseScale() == 1 {
		return 0
	}
	return value
}

func (w *wrapper) subShift(script *wrapper, defaultShift float64) float64 {
	core := w.scriptBaseCore()
	if core == nil || script == nil {
		return 0
	}
	bbox, sbox := core.outerBBox(), script.outerBBox()
	params := w.renderer.params
	shift := layout.Length2Em(attribute(w.node, "subscriptshift", ""), defaultShift, w.bbox.Scale, w.renderer.pxPerEm)
	return math.Max(
		w.baseCharZero(bbox.D*w.baseScale()+params.SubDrop*sbox.RScale),
		math.Max(shift, sbox.H*sbox.RScale-.8*params.XHeight),
	)
}

func (w *wrapper) supShift(script *wrapper) float64 {
	core := w.scriptBaseCore()
	if core == nil || script == nil {
		return 0
	}
	bbox, sbox := core.outerBBox(), script.outerBBox()
	params := w.renderer.params
	minimum := params.Sup2
	if value, ok := w.node.Property("texprimestyle"); ok && truthy(value) {
		minimum = params.Sup3
	} else if w.displayStyle {
		minimum = params.Sup1
	}
	shift := layout.Length2Em(attribute(w.node, "superscriptshift", ""), minimum, w.bbox.Scale, w.renderer.pxPerEm)
	return math.Max(
		w.baseCharZero(bbox.H*w.baseScale()-params.SupDrop*sbox.RScale),
		math.Max(shift, sbox.D*sbox.RScale+.25*params.XHeight),
	)
}

func (w *wrapper) scriptOffsets() (subY, supY float64) {
	if len(w.children) < 3 {
		return 0, 0
	}
	sub, sup := w.children[1], w.children[2]
	subbox, supbox := sub.outerBBox(), sup.outerBBox()
	params := w.renderer.params
	u := w.supShift(sup)
	drop := w.baseCharZero(w.scriptBaseCore().outerBBox().D*w.baseScale() + params.SubDrop*subbox.RScale)
	v := math.Max(drop, layout.Length2Em(attribute(w.node, "subscriptshift", ""), params.Sub2, w.bbox.Scale, w.renderer.pxPerEm))
	minimum := 3 * params.Rule
	space := (u - supbox.D*supbox.RScale) - (subbox.H*subbox.RScale - v)
	if space < minimum {
		v += minimum - space
		p := .8*params.XHeight - (u - supbox.D*supbox.RScale)
		if p > 0 {
			u += p
			v -= p
		}
	}
	u = math.Max(layout.Length2Em(attribute(w.node, "superscriptshift", ""), u, w.bbox.Scale, w.renderer.pxPerEm), u)
	v = math.Max(layout.Length2Em(attribute(w.node, "subscriptshift", ""), v, w.bbox.Scale, w.renderer.pxPerEm), v)
	return -v, u
}

func (w *wrapper) computeScriptsBBox(bbox *layout.BBox) {
	base := w.scriptBase()
	if base == nil || len(w.children) < 2 {
		w.computeChildrenBBox(bbox)
		return
	}
	bbox.Empty()
	bbox.Append(base.outerBBox())
	width := w.baseWidth()
	switch w.node.Kind {
	case "msub":
		script := w.children[1]
		bbox.Combine(script.outerBBox(), width, -w.subShift(script, w.renderer.params.Sub1))
	case "msup":
		script := w.children[1]
		x := w.adjustedIC() - w.baseIC()
		bbox.Combine(script.outerBBox(), width+x, w.supShift(script))
	case "msubsup":
		sub, sup := w.children[1], w.children[2]
		subY, supY := w.scriptOffsets()
		bbox.Combine(sub.outerBBox(), width, subY)
		bbox.Combine(sup.outerBBox(), width+w.adjustedIC(), supY)
	}
	bbox.W += w.renderer.params.ScriptSpace
	bbox.Clean()
}

func (w *wrapper) scriptsToSVG(parent *Element) {
	element := w.standardSVG(parent)
	base := w.scriptBase()
	if base == nil || len(w.children) < 2 {
		return
	}
	width := w.baseWidth()
	base.toSVG(element)
	switch w.node.Kind {
	case "msub":
		script := w.children[1]
		script.toSVG(element)
		script.place(width, -w.subShift(script, w.renderer.params.Sub1), script.element)
	case "msup":
		script := w.children[1]
		script.toSVG(element)
		x := w.adjustedIC() - w.baseIC()
		script.place(width+x, w.supShift(script), script.element)
	case "msubsup":
		sub, sup := w.children[1], w.children[2]
		subY, supY := w.scriptOffsets()
		// MathJax emits the superscript before the subscript.
		sup.toSVG(element)
		sub.toSVG(element)
		sub.place(width, subY, sub.element)
		sup.place(width+w.adjustedIC(), supY, sup.element)
	}
}
