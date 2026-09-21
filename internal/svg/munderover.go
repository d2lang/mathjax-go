// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
)

func (w *wrapper) underChild() *wrapper {
	if (w.node.Kind == "munder" || w.node.Kind == "munderover") && len(w.children) > 1 {
		return w.children[1]
	}
	return nil
}

func (w *wrapper) overChild() *wrapper {
	if w.node.Kind == "mover" && len(w.children) > 1 {
		return w.children[1]
	}
	if w.node.Kind == "munderover" && len(w.children) > 2 {
		return w.children[2]
	}
	return nil
}

// movableLimitCore follows the node-specific coreMO contract used only by
// CommonScriptbase.hasMovableLimits. A non-operator may be the returned core.
// Keep the distinct layout-base and accent-operator traversals unchanged.
func movableLimitCore(base *wrapper) *wrapper {
	for base != nil {
		index := 0
		switch base.node.Kind {
		case "mrow":
			if !base.node.Flags.Embellished {
				return base
			}
			index = base.node.Flags.CoreIndex
		case "msub", "msup", "msubsup", "munder", "mover", "munderover", "mmultiscripts",
			"TeXAtom", "mstyle", "mpadded", "mphantom", "semantics", "mfrac", "math", "mtd":
			// AbstractMmlBaseNode and AbstractMmlLayoutNode always use child 0.
		case "maction":
			selection, ok := numberAttribute(base.node, "selection")
			if !ok || math.IsNaN(selection) {
				return nil
			}
			selected := math.Max(1, math.Min(float64(len(base.children)), selection)) - 1
			if selected != math.Trunc(selected) {
				return nil
			}
			index = int(selected)
		default:
			return base
		}
		if index < 0 || index >= len(base.children) {
			return nil
		}
		base = base.children[index]
	}
	return nil
}

func (w *wrapper) hasMovableLimits() bool {
	base := movableLimitCore(w.scriptBase())
	return !w.displayStyle && base != nil && boolAttributeDefault(base.node, "movablelimits", false)
}

func (w *wrapper) lineAccent(script *wrapper) bool {
	return script != nil && nodeText(script.node) == "―"
}

// CommonScriptbase records the first under/over accent below transparent
// wrappers. Its existing gap is removed before a further label is stacked.
func (w *wrapper) baseHasAccent(attribute string) bool {
	core := w.scriptBaseCore()
	if core == nil {
		return false
	}
	switch core.node.Kind {
	case "munder", "mover", "munderover":
		return boolAttributeDefault(core.node, attribute, false)
	}
	return false
}

func (w *wrapper) overKU(base, over *layout.BBox) (separation, offset float64) {
	accent := boolAttributeDefault(w.node, "accent", false)
	params := w.renderer.params
	depth := over.D * over.RScale
	t := params.Rule * params.SeparationFactor
	T := t
	if w.lineAccent(w.overChild()) {
		T = 3 * params.Rule
	}
	separation = math.Max(params.BigOp1, params.BigOp3-math.Max(0, depth))
	if accent {
		separation = T
	}
	if w.baseHasAccent("accent") {
		separation -= t
	}
	offset = base.H*base.RScale + separation + depth
	return
}

func (w *wrapper) underKV(base, under *layout.BBox) (separation, offset float64) {
	accent := boolAttributeDefault(w.node, "accentunder", false)
	params := w.renderer.params
	height := under.H * under.RScale
	t := params.Rule * params.SeparationFactor
	T := t
	if w.lineAccent(w.underChild()) {
		T = 3 * params.Rule
	}
	separation = math.Max(params.BigOp2, params.BigOp4-height)
	if accent {
		separation = T
	}
	if w.baseHasAccent("accentunder") {
		separation -= t
	}
	offset = -(base.D*base.RScale + separation + height)
	return
}

func (w *wrapper) underOverDelta(noSkew bool) float64 {
	core := w.scriptBaseCore()
	if core == nil {
		return 0
	}
	bbox := core.outerBBox()
	skew := 0.0
	if boolAttributeDefault(w.node, "accent", false) && !noSkew {
		skew = bbox.Skew
	}
	return (skew + .75*bbox.IC) * w.baseScale()
}

func (w *wrapper) stackOffsets(boxes []*layout.BBox, deltas []float64) []float64 {
	widths := make([]float64, len(boxes))
	for i, box := range boxes {
		widths[i] = box.W * box.RScale
	}
	base := w.scriptBaseCore()
	if len(widths) != 0 && w.removeBaseIC() && base != nil && !boolAttributeDefault(base.node, "largeop", false) {
		widths[0] -= w.baseIC()
	}
	width := 0.0
	for _, value := range widths {
		width = math.Max(width, value)
	}
	align := stringAttribute(w.node, "align", "center")
	offsets := make([]float64, len(boxes))
	minimum := 0.0
	for i := range boxes {
		switch align {
		case "right":
			offsets[i] = width - widths[i]
		case "left":
			offsets[i] = 0
		default:
			offsets[i] = (width - widths[i]) / 2
		}
		if i < len(deltas) {
			offsets[i] += deltas[i]
		}
		if offsets[i] < minimum {
			minimum = -offsets[i]
		}
	}
	if minimum != 0 {
		for i := range offsets {
			offsets[i] += minimum
		}
	}
	for i := 1; i < len(offsets); i++ {
		offsets[i] += boxes[i].DX * boxes[0].Scale
	}
	return offsets
}

type underOverStretchChild struct {
	child *wrapper
	core  *wrapper
}

// stretchUnderOverChildren ports CommonScriptbase.stretchChildren().  The
// child wrapper participates in width selection, while its embellished core
// mo owns the delimiter variant.  Wide accents and x-arrows therefore select
// the same MathJax size glyph before the stack is measured.
func (w *wrapper) stretchUnderOverChildren() {
	stretchable := make([]underOverStretchChild, 0, len(w.children))
	for _, child := range w.children {
		core := child
		for core != nil && core.node.Kind != "mo" && core.node.Flags.Embellished {
			index := core.node.Flags.CoreIndex
			if index < 0 || index >= len(core.children) {
				core = nil
				break
			}
			core = core.children[index]
		}
		if core != nil && core.canStretch(font.DirectionHorizontal) {
			stretchable = append(stretchable, underOverStretchChild{child: child, core: core})
		}
	}
	if len(stretchable) == 0 || len(w.children) <= 1 {
		return
	}

	all := len(stretchable) > 1 && len(stretchable) == len(w.children)
	width := 0.0
	for _, child := range w.children {
		isStretchable := false
		for _, candidate := range stretchable {
			if candidate.child == child {
				isStretchable = true
				break
			}
		}
		if !all && isStretchable {
			continue
		}
		bbox := child.outerBBox()
		width = math.Max(width, bbox.W*bbox.RScale)
	}
	for _, child := range stretchable {
		scale := child.child.bbox.RScale
		if scale == 0 {
			scale = 1
		}
		child.core.getStretchedVariant([]float64{width / scale}, false)
	}
}

func (w *wrapper) computeMovableLimitsBBox(bbox *layout.BBox) {
	base := w.scriptBase()
	if base == nil {
		w.computeChildrenBBox(bbox)
		return
	}
	bbox.Empty()
	bbox.Append(base.outerBBox())
	width := w.baseWidth()
	switch w.node.Kind {
	case "munder":
		sub := w.underChild()
		bbox.Combine(sub.outerBBox(), width, -w.subShift(sub, w.renderer.params.Sub1))
	case "mover":
		sup := w.overChild()
		x := w.adjustedIC() - w.baseIC()
		bbox.Combine(sup.outerBBox(), width+x, w.supShift(sup))
	case "munderover":
		sub, sup := w.underChild(), w.overChild()
		subY, supY := w.scriptOffsetsFor(sub, sup)
		bbox.Combine(sub.outerBBox(), width, subY)
		bbox.Combine(sup.outerBBox(), width+w.adjustedIC(), supY)
	}
	bbox.W += w.renderer.params.ScriptSpace
	bbox.Clean()
}

func (w *wrapper) scriptOffsetsFor(sub, sup *wrapper) (subY, supY float64) {
	if sub == nil || sup == nil {
		return 0, 0
	}
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
	return -v, u
}

func (w *wrapper) computeUnderOverBBox(bbox *layout.BBox) {
	if w.hasMovableLimits() {
		w.computeMovableLimitsBBox(bbox)
		return
	}
	base := w.scriptBase()
	if base == nil {
		w.computeChildrenBBox(bbox)
		return
	}
	bbox.Empty()
	baseBox := base.outerBBox()
	if boolAttributeDefault(w.node, "accent", false) {
		baseBox.H = math.Max(baseBox.H, w.renderer.params.XHeight*baseBox.Scale)
	}
	switch w.node.Kind {
	case "munder":
		under := w.underChild()
		underBox := under.outerBBox()
		_, v := w.underKV(baseBox, underBox)
		delta := w.underOverDelta(true)
		if w.lineAccent(under) {
			delta = 0
		}
		x := w.stackOffsets([]*layout.BBox{baseBox, underBox}, []float64{0, -delta})
		bbox.Combine(baseBox, x[0], 0)
		bbox.Combine(underBox, x[1], v)
		bbox.D += w.renderer.params.BigOp5
	case "mover":
		over := w.overChild()
		overBox := over.outerBBox()
		_, u := w.overKU(baseBox, overBox)
		delta := w.underOverDelta(false)
		if w.lineAccent(over) {
			delta = 0
		}
		x := w.stackOffsets([]*layout.BBox{baseBox, overBox}, []float64{0, delta})
		bbox.Combine(baseBox, x[0], 0)
		bbox.Combine(overBox, x[1], u)
		bbox.H += w.renderer.params.BigOp5
	case "munderover":
		under, over := w.underChild(), w.overChild()
		underBox, overBox := under.outerBBox(), over.outerBBox()
		_, u := w.overKU(baseBox, overBox)
		_, v := w.underKV(baseBox, underBox)
		delta := w.underOverDelta(false)
		underDelta, overDelta := -delta, delta
		if w.lineAccent(under) {
			underDelta = 0
		}
		if w.lineAccent(over) {
			overDelta = 0
		}
		x := w.stackOffsets([]*layout.BBox{baseBox, underBox, overBox}, []float64{0, underDelta, overDelta})
		bbox.Combine(baseBox, x[0], 0)
		bbox.Combine(overBox, x[2], u)
		bbox.Combine(underBox, x[1], v)
		bbox.H += w.renderer.params.BigOp5
		bbox.D += w.renderer.params.BigOp5
	}
	bbox.Clean()
}

func (w *wrapper) underOverToSVG(parent *Element) {
	if w.hasMovableLimits() {
		w.movableLimitsToSVG(parent)
		return
	}
	element := w.standardSVG(parent)
	base := w.scriptBase()
	if base == nil {
		return
	}
	baseBox := base.outerBBox()
	base.toSVG(element)
	switch w.node.Kind {
	case "munder":
		under := w.underChild()
		underBox := under.outerBBox()
		under.toSVG(element)
		delta := w.underOverDelta(true)
		if w.lineAccent(under) {
			delta = 0
		}
		_, v := w.underKV(baseBox, underBox)
		x := w.stackOffsets([]*layout.BBox{baseBox, underBox}, []float64{0, -delta})
		base.place(x[0], 0, base.element)
		under.place(x[1], v, under.element)
	case "mover":
		over := w.overChild()
		overBox := over.outerBBox()
		over.toSVG(element)
		delta := w.underOverDelta(false)
		if w.lineAccent(over) {
			delta = 0
		}
		_, u := w.overKU(baseBox, overBox)
		x := w.stackOffsets([]*layout.BBox{baseBox, overBox}, []float64{0, delta})
		base.place(x[0], 0, base.element)
		over.place(x[1], u, over.element)
	case "munderover":
		under, over := w.underChild(), w.overChild()
		underBox, overBox := under.outerBBox(), over.outerBBox()
		under.toSVG(element)
		over.toSVG(element)
		delta := w.underOverDelta(false)
		underDelta, overDelta := -delta, delta
		if w.lineAccent(under) {
			underDelta = 0
		}
		if w.lineAccent(over) {
			overDelta = 0
		}
		_, u := w.overKU(baseBox, overBox)
		_, v := w.underKV(baseBox, underBox)
		x := w.stackOffsets([]*layout.BBox{baseBox, underBox, overBox}, []float64{0, underDelta, overDelta})
		base.place(x[0], 0, base.element)
		under.place(x[1], v, under.element)
		over.place(x[2], u, over.element)
	}
}

func (w *wrapper) movableLimitsToSVG(parent *Element) {
	element := w.standardSVG(parent)
	base := w.scriptBase()
	base.toSVG(element)
	width := w.baseWidth()
	switch w.node.Kind {
	case "munder":
		sub := w.underChild()
		sub.toSVG(element)
		sub.place(width, -w.subShift(sub, w.renderer.params.Sub1), sub.element)
	case "mover":
		sup := w.overChild()
		sup.toSVG(element)
		sup.place(width+w.adjustedIC()-w.baseIC(), w.supShift(sup), sup.element)
	case "munderover":
		sub, sup := w.underChild(), w.overChild()
		subY, supY := w.scriptOffsetsFor(sub, sup)
		sup.toSVG(element)
		sub.toSVG(element)
		sub.place(width, subY, sub.element)
		sup.place(width+w.adjustedIC(), supY, sup.element)
	}
}
