// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/output/common/Wrappers/menclose.ts and
// ts/output/svg/Wrappers/menclose.ts.

package svg

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

const (
	encloseArrowX  = 4.0
	encloseArrowDX = 1.0
	encloseArrowY  = 2.0
	encloseRule    = .067
	enclosePadding = .2
)

type encloseTRBL [4]float64

type encloseLayout struct {
	w         *wrapper
	padding   float64
	thickness float64
	arrowX    float64
	arrowY    float64
	arrowDX   float64
	notations []string
	active    map[string]bool
	trbl      encloseTRBL
	msqrt     *wrapper
}

var encloseFloatPrefix = regexp.MustCompile(`^[+-]?(?:Infinity|(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)`)

func newEncloseLayout(w *wrapper) *encloseLayout {
	layoutData := &encloseLayout{
		w: w, padding: enclosePadding, thickness: encloseRule,
		arrowX: encloseArrowX, arrowY: encloseArrowY, arrowDX: encloseArrowDX,
		active: make(map[string]bool),
	}
	layoutData.getParameters()
	layoutData.getNotations()
	layoutData.removeRedundantNotations()
	if layoutData.active["radical"] {
		layoutData.msqrt = layoutData.createMsqrt()
	}
	layoutData.trbl = layoutData.getBBoxExtenders()
	return layoutData
}

func (e *encloseLayout) getParameters() {
	if value, ok := e.w.node.Attributes.Get("data-padding"); ok {
		e.padding = layout.Length2Em(value, enclosePadding, e.w.bbox.Scale, e.w.renderer.pxPerEm)
	}
	if value, ok := e.w.node.Attributes.Get("data-thickness"); ok {
		e.thickness = layout.Length2Em(value, encloseRule, e.w.bbox.Scale, e.w.renderer.pxPerEm)
	}
	if value, ok := e.w.node.Attributes.Get("data-arrowhead"); ok {
		parts := strings.Fields(encloseStringValue(value))
		e.arrowX = encloseArrowPart(parts, 0, encloseArrowX)
		e.arrowY = encloseArrowPart(parts, 1, encloseArrowY)
		e.arrowDX = encloseArrowPart(parts, 2, encloseArrowDX)
	}
}

func encloseStringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return strconv.FormatFloat(encloseNumberValue(value), 'g', -1, 64)
}

func encloseNumberValue(value any) float64 {
	switch value := value.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		return encloseParseFloat(value)
	default:
		return math.NaN()
	}
}

func encloseArrowPart(parts []string, index int, fallback float64) float64 {
	if index >= len(parts) || parts[index] == "" {
		return fallback
	}
	return encloseParseFloat(parts[index])
}

func encloseParseFloat(value string) float64 {
	match := encloseFloatPrefix.FindString(strings.TrimSpace(value))
	if match == "" {
		return math.NaN()
	}
	parsed, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return math.NaN()
	}
	return parsed
}

func (e *encloseLayout) getNotations() {
	notation := stringAttribute(e.w.node, "notation", "longdiv")
	for _, name := range strings.Fields(notation) {
		if !encloseNotationSupported(name) || e.active[name] {
			continue
		}
		e.active[name] = true
		e.notations = append(e.notations, name)
	}
}

func (e *encloseLayout) removeRedundantNotations() {
	for _, name := range e.notations {
		if !e.active[name] {
			continue
		}
		for _, redundant := range strings.Fields(encloseNotationRemove(name)) {
			delete(e.active, redundant)
		}
	}
}

func (e *encloseLayout) activeNotations() []string {
	active := make([]string, 0, len(e.notations))
	for _, name := range e.notations {
		if e.active[name] {
			active = append(active, name)
		}
	}
	return active
}

func (e *encloseLayout) getBBoxExtenders() encloseTRBL {
	var trbl encloseTRBL
	for _, name := range e.activeNotations() {
		maximizeEnclose(&trbl, e.notationBBox(name))
	}
	return trbl
}

func (e *encloseLayout) getPadding() encloseTRBL {
	var border encloseTRBL
	for _, name := range e.activeNotations() {
		maximizeEnclose(&border, e.notationBorder(name))
	}
	padding := e.trbl
	for i := range padding {
		padding[i] -= border[i]
	}
	return padding
}

func maximizeEnclose(target *encloseTRBL, values encloseTRBL) {
	for i, value := range values {
		if (*target)[i] < value {
			(*target)[i] = value
		}
	}
}

func (e *encloseLayout) getOffset(direction string) float64 {
	value := (e.trbl[2] - e.trbl[0]) / 2
	if direction == "X" {
		value = (e.trbl[1] - e.trbl[3]) / 2
	}
	if math.Abs(value) <= .001 {
		return 0
	}
	return value
}

func encloseArgMod(width, height float64) (float64, float64) {
	return math.Atan2(height, width), math.Sqrt(width*width + height*height)
}

func (e *encloseLayout) arrowData() (angle, width, x, y float64) {
	r := e.thickness * (e.arrowX + math.Max(1, e.arrowDX))
	child := e.childBBox()
	height := child.H + child.D
	radius := math.Sqrt(height*height + child.W*child.W)
	x = math.Max(e.padding, r*child.W/radius)
	y = math.Max(e.padding, r*height/radius)
	angle, width = encloseArgMod(child.W+2*x, height+2*y)
	return
}

func (e *encloseLayout) arrowAW() (float64, float64) {
	child := e.childBBox()
	return encloseArgMod(
		e.trbl[3]+child.W+e.trbl[1],
		e.trbl[0]+child.H+child.D+e.trbl[2],
	)
}

func (e *encloseLayout) childBBox() *layout.BBox {
	if len(e.w.children) == 0 {
		return layout.ZeroBBox()
	}
	return e.w.children[0].getBBox()
}

func (e *encloseLayout) createMsqrt() *wrapper {
	if len(e.w.node.Children) == 0 {
		return nil
	}
	root := mml.NewNode("msqrt", nil, nil, e.w.node.Children[0].Clone())
	root.TeXClass = mml.TeXClassOrd
	root.Attributes.SetInherited("displaystyle", e.w.displayStyle)
	root.Attributes.SetInherited("scriptlevel", e.w.scriptLevel)
	if value, ok := e.w.node.Attributes.Get("mathsize"); ok && e.w.node.Attributes.IsSet("mathsize") {
		root.Attributes.SetInherited("mathsize", value)
	}
	return e.w.renderer.wrap(root, e.w, e.w.scriptLevel, e.w.displayStyle)
}

func (e *encloseLayout) sqrtTRBL() encloseTRBL {
	if e.msqrt == nil || len(e.msqrt.children) == 0 {
		return encloseTRBL{}
	}
	root := e.msqrt.getBBox()
	child := e.msqrt.children[0].getBBox()
	return encloseTRBL{root.H - child.H, 0, root.D - child.D, root.W - child.W}
}

// computeEncloseBBox ports CommonMencloseMixin.computeBBox().  It is named for
// the central wrapper dispatcher and deliberately keeps all notation state
// local so wrappers remain immutable and safe for concurrent renders.
func (w *wrapper) computeEncloseBBox(bbox *layout.BBox) {
	state := newEncloseLayout(w)
	bbox.Empty()
	bbox.Combine(state.childBBox(), state.trbl[3], 0)
	bbox.H += state.trbl[0]
	bbox.D += state.trbl[2]
	bbox.W += state.trbl[1]
	bbox.Clean()
}

// encloseToSVG ports SVGmenclose.toSVG().
func (w *wrapper) encloseToSVG(parent *Element) {
	if !w.bboxComputed {
		w.computeEncloseBBox(w.bbox)
		w.bboxComputed = true
	}
	state := newEncloseLayout(w)
	element := w.standardSVG(parent)
	block := NewElement("g")
	if left := state.getBBoxExtenders()[3]; left > 0 {
		block.SetAttr("transform", "translate("+fixed(left)+", 0)")
	}
	element.Append(block)
	if state.active["radical"] {
		state.renderRadical(block)
	} else if len(w.children) != 0 {
		w.children[0].toSVG(block)
	}
	for _, name := range state.activeNotations() {
		if name != "radical" {
			state.renderNotation(name, element)
		}
	}
}

func (e *encloseLayout) renderRadical(block *Element) {
	if e.msqrt == nil {
		return
	}
	e.msqrt.rootToSVG(block)
	e.w.place(-e.sqrtTRBL()[3], 0, block)
}
