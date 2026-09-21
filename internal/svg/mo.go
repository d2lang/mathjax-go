// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
)

// defaultAccentRemap is FontData.defaultAccentMap from MathJax 3.2.2.
// CommonMo applies it only to a single-character accent token outside an
// mrow.  Several entries deliberately expand to more than one character.
var defaultAccentRemap = map[rune]string{
	0x0300: "\u02CB", 0x0301: "\u02CA", 0x0302: "\u02C6", 0x0303: "\u02DC",
	0x0304: "\u02C9", 0x0306: "\u02D8", 0x0307: "\u02D9", 0x0308: "\u00A8",
	0x030A: "\u02DA", 0x030C: "\u02C7", 0x2192: "\u20D7",
	0x2032: "'", 0x2033: "''", 0x2034: "'''", 0x2035: "`",
	0x2036: "``", 0x2037: "```", 0x2057: "''''",
	0x20D0: "\u21BC", 0x20D1: "\u21C0", 0x20D6: "\u2190", 0x20E1: "\u2194",
	0x20F0: "*", 0x20DB: "...", 0x20DC: "....",
	0x20EC: "\u21C1", 0x20ED: "\u21BD", 0x20EE: "\u2190", 0x20EF: "\u2192",
}

func accentCoreMO(w *wrapper) *wrapper {
	for w != nil && w.node.Kind != "mo" && w.node.Flags.Embellished {
		index := w.node.Flags.CoreIndex
		if index < 0 || index >= len(w.children) {
			return nil
		}
		w = w.children[index]
	}
	if w != nil && w.node.Kind == "mo" {
		return w
	}
	return nil
}

func (w *wrapper) isAccentMO() bool {
	if w == nil || w.node.Kind != "mo" {
		return false
	}
	// MmlMo.isAccent is separate from the internal mathaccent positioning
	// marker. Follow coreParent through embellished nodes, stopping at math.
	outer := w
	for outer.parent != nil && outer.parent.node.Kind != "math" &&
		outer.parent.node.Flags.Embellished && accentCoreMO(outer.parent) == w {
		outer = outer.parent
	}
	stack := outer.parent
	if stack == nil {
		return false
	}
	key := ""
	switch stack.node.Kind {
	case "mover":
		if stack.overChild() != nil {
			key = "accent"
		}
	case "munder":
		if stack.underChild() != nil {
			key = "accentunder"
		}
	case "munderover":
		if accentCoreMO(stack.overChild()) == w {
			key = "accent"
		} else if accentCoreMO(stack.underChild()) == w {
			key = "accentunder"
		}
	}
	if key == "" {
		return false
	}
	// The pinned getter leaves isAccent false when the stack has an explicit
	// accent attribute; otherwise it uses this operator's resolved accent.
	if _, explicit := stack.node.Attributes.GetExplicit(key); explicit {
		return false
	}
	return boolAttributeDefault(w.node, "accent", false)
}

func remapAccentText(parent *wrapper, text string) string {
	// CommonTextNode uses an explicit stretch.c directly. A selected schar
	// also becomes stretch.c; neither may be remapped back to an accent.
	if parent != nil && parent.hasStretch {
		hasCharacter := parent.stretch.HasAlias && parent.stretch.Alias != 0
		if parent.sizeSet && parent.size >= 0 && parent.size < len(parent.stretch.SizeChars) {
			hasCharacter = hasCharacter || parent.stretch.SizeChars[parent.size] != 0
		}
		if hasCharacter {
			return text
		}
	}
	characters := []rune(text)
	if len(characters) != 1 || parent == nil || !parent.isAccentMO() {
		return text
	}
	if remapped, ok := defaultAccentRemap[characters[0]]; ok {
		return remapped
	}
	return text
}

func (w *wrapper) canStretch(direction font.Direction) bool {
	if w.node.Kind != "mo" {
		// CommonWrapper.canStretch resets its own wrapper-level stretch state,
		// then delegates through the core child of an embellished wrapper.  The
		// wrapper state is observable when a table later asks a mover/mrow to
		// stretch after its core mo was already sized by an earlier layout pass.
		w.stretch = font.Delimiter{}
		w.hasStretch = false
		if !w.node.Flags.Embellished {
			return false
		}
		index := w.node.Flags.CoreIndex
		if index < 0 || index >= len(w.children) {
			return false
		}
		core := w.children[index]
		if core == nil || core == w || !core.canStretch(direction) {
			return false
		}
		w.stretch = core.stretch
		w.hasStretch = true
		return true
	}
	if w.hasStretch {
		return w.stretch.Direction == direction
	}
	if !boolAttributeDefault(w.node, "stretchy", false) {
		return false
	}
	text := []rune(nodeText(w.node))
	if len(text) != 1 {
		return false
	}
	delimiter, ok := font.LookupDelimiter(text[0])
	if !ok || delimiter.Direction != direction {
		return false
	}
	w.stretch = delimiter
	w.hasStretch = true
	w.stretchGlyph = text[0]
	return true
}

func (w *wrapper) stretchDimension(dimensions []float64) float64 {
	if len(dimensions) == 0 {
		return 0
	}
	if len(dimensions) == 1 {
		return dimensions[0]
	}
	height, depth := dimensions[0], dimensions[1]
	if boolAttributeDefault(w.node, "symmetric", false) {
		a := w.renderer.params.Axis
		return 2 * math.Max(height-a, depth+a)
	}
	return height + depth
}

func (w *wrapper) delimiterSize(name string, fallback float64) float64 {
	if w.node.Attributes.IsSet(name) {
		return layout.Length2Em(attribute(w.node, name, fallback), 1, 1, w.renderer.pxPerEm)
	}
	return fallback
}

func (w *wrapper) getStretchedVariant(dimensions []float64, exact bool) {
	if !w.hasStretch {
		return
	}
	dimension := w.stretchDimension(dimensions)
	minimum := w.delimiterSize("minsize", 0)
	maximum := w.delimiterSize("maxsize", math.Inf(1))
	dimension = math.Max(minimum, math.Min(maximum, dimension))
	factor := w.renderer.params.DelimiterFactor / 1000
	shortfall := w.renderer.params.DelimiterShortfall
	target := dimension
	_, mathAccent := w.node.Property("mathaccent")
	if minimum == 0 && !exact {
		if mathAccent {
			target = math.Min(dimension/factor, dimension+shortfall)
		} else {
			target = math.Max(dimension*factor, dimension-shortfall)
		}
	}
	for i := range w.stretch.Sizes {
		if w.stretch.Sizes[i] < target {
			continue
		}
		if mathAccent && i != 0 {
			i--
		}
		if variant, ok := w.stretch.SizeVariant(i); ok {
			w.variant = variant
		}
		w.size, w.sizeSet = i, true
		if i < len(w.stretch.SizeChars) && w.stretch.SizeChars[i] != 0 {
			w.stretchGlyph = w.stretch.SizeChars[i]
		} else if w.stretch.HasAlias {
			w.stretchGlyph = w.stretch.Alias
		}
		w.invalidateBBox()
		return
	}
	if len(w.stretch.Stretch) != 0 {
		w.size, w.sizeSet = -1, true
		w.setStretchBBox(dimensions, w.extendedHeight(dimension))
		return
	}
	if len(w.stretch.Sizes) != 0 {
		i := len(w.stretch.Sizes) - 1
		if variant, ok := w.stretch.SizeVariant(i); ok {
			w.variant = variant
		}
		w.size, w.sizeSet = i, true
		// CommonMo retains delim.c when falling back to the largest fixed size.
		if w.stretch.HasAlias {
			w.stretchGlyph = w.stretch.Alias
		}
		w.invalidateBBox()
	}
}

func (w *wrapper) extendedHeight(dimension float64) float64 {
	if w.stretch.HasFullExt {
		extender, ends := w.stretch.FullExt[0], w.stretch.FullExt[1]
		count := math.Ceil(math.Max(0, dimension-ends) / extender)
		return ends + count*extender
	}
	return dimension
}

func (w *wrapper) setStretchBBox(dimensions []float64, dimension float64) {
	if w.stretch.HasMin && w.stretch.Min > dimension {
		dimension = w.stretch.Min
	}
	height, depth, width := w.stretch.HDW[0], w.stretch.HDW[1], w.stretch.HDW[2]
	if w.stretch.Direction == font.DirectionVertical {
		height, depth = w.stretchBaseline(dimensions, dimension)
	} else {
		width = dimension
	}
	w.bbox.H, w.bbox.D, w.bbox.W = height, depth, width
	w.bboxComputed = true
}

func (w *wrapper) stretchBaseline(dimensions []float64, total float64) (float64, float64) {
	hasHeightDepth := len(dimensions) == 2 && dimensions[0]+dimensions[1] == total
	height, depth := total, 0.0
	if boolAttributeDefault(w.node, "symmetric", false) {
		a := w.renderer.params.Axis
		if hasHeightDepth {
			height = 2 * math.Max(dimensions[0]-a, dimensions[1]+a)
		}
		depth = height/2 - a
	} else if hasHeightDepth {
		depth = dimensions[1]
	} else {
		ch, cd := .75, .25
		if w.stretch.HasHDW {
			ch, cd = w.stretch.HDW[0], w.stretch.HDW[1]
		}
		depth = cd * height / (ch + cd)
	}
	return height - depth, depth
}

func (w *wrapper) invalidateBBox() {
	w.bboxComputed = false
	for _, child := range w.children {
		child.bboxComputed = false
	}
	if w.parent != nil && w.parent.bboxComputed {
		w.parent.invalidateBBox()
	}
}

func (w *wrapper) computeMoBBox(bbox *layout.BBox) {
	if w.hasStretch && !w.sizeSet {
		w.getStretchedVariant([]float64{0}, false)
	}
	if !(w.hasStretch && w.sizeSet && w.size < 0) {
		w.computeChildrenBBox(bbox)
		w.copySkewIC(bbox)
	}
	if boolAttributeDefault(w.node, "symmetric", false) && (!w.hasStretch || w.stretch.Direction != font.DirectionHorizontal) {
		offset := w.centerOffset(bbox)
		bbox.H += offset
		bbox.D -= offset
	}
	if _, accent := w.node.Property("mathaccent"); accent && (!w.hasStretch || w.size >= 0) {
		bbox.W = 0
	}
}

func (w *wrapper) centerOffset(bbox *layout.BBox) float64 {
	return (bbox.H+bbox.D)/2 + w.renderer.params.Axis - bbox.H
}

func (w *wrapper) moToSVG(parent *Element) {
	element := w.standardSVG(parent)
	if w.hasStretch && !w.sizeSet {
		w.getStretchedVariant(nil, false)
	}
	if w.hasStretch && w.sizeSet && w.size < 0 {
		w.stretchToSVG(element)
		return
	}
	u, v := 0.0, 0.0
	if boolAttributeDefault(w.node, "symmetric", false) || boolAttributeDefault(w.node, "largeop", false) {
		prototype := layout.EmptyBBox()
		w.computeChildrenBBox(prototype)
		u = w.centerOffset(prototype)
	}
	if _, accent := w.node.Property("mathaccent"); accent {
		prototype := layout.EmptyBBox()
		w.computeChildrenBBox(prototype)
		// CommonMo.getAccentOffset uses protoBBox, including italic correction.
		w.copySkewIC(prototype)
		v = -prototype.W / 2
	}
	if u != 0 || v != 0 {
		element.SetAttr("transform", "translate("+fixed(v)+" "+fixed(u)+")")
	}
	w.addChildren(element)
}

func (w *wrapper) stretchToSVG(element *Element) {
	if w.stretch.Direction == font.DirectionVertical {
		w.stretchVertical(element)
	} else {
		w.stretchHorizontal(element)
	}
}

func (w *wrapper) stretchRowChildren() {
	var stretchable []*wrapper
	for _, child := range w.children {
		core := child
		for core != nil && core.node.Flags.Embellished && core.node.Kind != "mo" && len(core.children) != 0 {
			index := core.node.Flags.CoreIndex
			if index < 0 || index >= len(core.children) {
				index = 0
			}
			core = core.children[index]
		}
		if core != nil && core.canStretch(font.DirectionVertical) {
			stretchable = append(stretchable, core)
		}
	}
	if len(stretchable) == 0 || len(w.children) <= 1 {
		return
	}
	height, depth := 0.0, 0.0
	all := len(stretchable) > 1 && len(stretchable) == len(w.children)
	for _, child := range w.children {
		isStretch := false
		for _, candidate := range stretchable {
			if candidate == child {
				isStretch = true
				break
			}
		}
		if !all && isStretch {
			continue
		}
		bbox := child.outerBBox()
		height = math.Max(height, bbox.H*bbox.RScale)
		depth = math.Max(depth, bbox.D*bbox.RScale)
	}
	for _, child := range stretchable {
		child.getStretchedVariant([]float64{height, depth}, false)
	}
}
