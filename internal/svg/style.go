// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"regexp"
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
)

const borderFuzz = .005

var cssCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)
var cssBorderWidthPattern = regexp.MustCompile(`^(?:[\d.]+[a-z]+|thin|medium|thick|inherit|initial|unset)$`)

type cssDeclaration struct {
	name  string
	value string
}

type borderStyle struct {
	raw   string
	width string
	style string
	color string
	set   bool
}

type wrapperStyles struct {
	border  [4]borderStyle
	padding [4]string
	padSet  [4]bool
	other   []cssDeclaration
}

var sideNames = [...]string{"top", "right", "bottom", "left"}

func (w *wrapper) initializeStyles() {
	style := stringAttributeExplicit(w.node, "style")
	if style == "" {
		return
	}
	w.styles = parseWrapperStyles(style)
	// The primary wrapper removes CSS font-size before rendering paths or
	// fallback text; getScale applies the value to the wrapper's geometry.
	kept := w.styles.other[:0]
	for _, declaration := range w.styles.other {
		switch declaration.name {
		case "font-size":
			w.cssFontSize = declaration.value
		case "font-weight":
			w.cssFontWeight = declaration.value
		case "font-style":
			w.cssFontStyle = declaration.value
		default:
			kept = append(kept, declaration)
		}
	}
	w.styles.other = kept
}

func parseWrapperStyles(source string) *wrapperStyles {
	styles := &wrapperStyles{}
	for _, declaration := range splitCSSDeclarations(cssCommentPattern.ReplaceAllString(source, "")) {
		name, value, ok := strings.Cut(declaration, ":")
		if !ok {
			continue
		}
		name = strings.ToLower(strings.TrimSpace(name))
		value = strings.TrimSpace(value)
		if name == "" {
			continue
		}
		if styles.setBorderComponent(name, value) {
			continue
		}
		if value == "" {
			continue
		}
		switch name {
		case "border":
			border := splitBorder(value)
			for i := range styles.border {
				styles.border[i] = border
			}
		case "border-top", "border-right", "border-bottom", "border-left":
			styles.border[sideIndex(strings.TrimPrefix(name, "border-"))] = splitBorder(value)
		case "border-width", "border-style", "border-color":
			values := splitTRBL(value)
			for i := range styles.border {
				styles.border[i].set = true
				styles.border[i].raw = ""
				switch strings.TrimPrefix(name, "border-") {
				case "width":
					styles.border[i].width = values[i]
				case "style":
					styles.border[i].style = values[i]
				case "color":
					styles.border[i].color = values[i]
				}
			}
		case "padding":
			values := splitTRBL(value)
			for i := range styles.padding {
				styles.padding[i], styles.padSet[i] = values[i], true
			}
		case "padding-top", "padding-right", "padding-bottom", "padding-left":
			i := sideIndex(strings.TrimPrefix(name, "padding-"))
			styles.padding[i], styles.padSet[i] = value, true
		default:
			styles.setOther(name, value)
		}
	}
	return styles
}

// Styles.set recombines a side's width/style/color after a component changes,
// including an empty value that removes a previously authored component.
func (s *wrapperStyles) setBorderComponent(name, value string) bool {
	parts := strings.Split(name, "-")
	if len(parts) != 3 || parts[0] != "border" {
		return false
	}
	i := -1
	for index, side := range sideNames {
		if parts[1] == side {
			i = index
			break
		}
	}
	if i < 0 {
		return false
	}
	border := &s.border[i]
	switch parts[2] {
	case "width":
		border.width = value
	case "style":
		border.style = value
	case "color":
		border.color = value
	default:
		return false
	}
	border.raw = ""
	border.set = border.cssValue() != ""
	return true
}

func splitCSSDeclarations(source string) []string {
	var declarations []string
	start := 0
	quote := rune(0)
	for index, character := range source {
		switch {
		case quote != 0 && character == quote:
			quote = 0
		case quote == 0 && (character == '\'' || character == '"'):
			quote = character
		case quote == 0 && character == ';':
			declarations = append(declarations, source[start:index])
			start = index + 1
		}
	}
	declarations = append(declarations, source[start:])
	return declarations
}

func splitBorder(value string) borderStyle {
	// Styles.splitWSC retains the authored shorthand until an individual
	// width/style/color component is changed and combineWSC rebuilds it.
	border := borderStyle{set: true, raw: value}
	for _, part := range splitCSSValue(value) {
		switch {
		case isBorderWidth(part) && border.width == "":
			border.width = part
		case isBorderStyle(part) && border.style == "":
			border.style = part
		default:
			border.color = part
		}
	}
	return border
}

func isBorderWidth(value string) bool {
	return cssBorderWidthPattern.MatchString(value)
}

func isBorderStyle(value string) bool {
	switch value {
	case "none", "hidden", "dotted", "dashed", "solid", "double", "groove", "ridge", "inset", "outset", "inherit", "initial", "unset":
		return true
	default:
		return false
	}
}

func splitTRBL(value string) [4]string {
	parts := splitCSSValue(value)
	if len(parts) == 0 {
		return [4]string{}
	}
	for len(parts) < 4 {
		switch len(parts) {
		case 1:
			parts = append(parts, parts[0])
		case 2:
			parts = append(parts, parts[0])
		case 3:
			parts = append(parts, parts[1])
		}
	}
	return [4]string{parts[0], parts[1], parts[2], parts[3]}
}

// Styles.splitSpaces keeps quoted strings and comma-followed whitespace in
// one part, so rgb()/hsl() colours remain intact in border shorthands. Repeated
// separating whitespace produces empty parts, as in the source regular expression.
func splitCSSValue(value string) []string {
	characters := []rune(value)
	var parts []string
	start := 0
	for i := 0; i < len(characters); i++ {
		c := characters[i]
		if c == '\'' || c == '"' {
			for j := i + 1; j < len(characters); j++ {
				if characters[j] == c {
					i = j
					break
				}
			}
			continue
		}
		if c == ',' && i+1 < len(characters) && cssValueSpace(characters[i+1]) {
			i++
			continue
		}
		if cssValueSpace(c) {
			parts = append(parts, string(characters[start:i]))
			start = i + 1
		}
	}
	if start < len(characters) {
		parts = append(parts, string(characters[start:]))
	}
	return parts
}

func cssValueSpace(c rune) bool {
	return c >= '\t' && c <= '\r' || c == ' ' || c == '\u00a0' || c == '\u1680' ||
		c >= '\u2000' && c <= '\u200a' || c == '\u2028' || c == '\u2029' ||
		c == '\u202f' || c == '\u205f' || c == '\u3000' || c == '\ufeff'
}

func sideIndex(name string) int {
	for i, side := range sideNames {
		if name == side {
			return i
		}
	}
	return 0
}

func (s *wrapperStyles) setOther(name, value string) {
	for i := range s.other {
		if s.other[i].name == name {
			s.other[i].value = value
			return
		}
	}
	s.other = append(s.other, cssDeclaration{name: name, value: value})
}

func (s *wrapperStyles) value(name string) string {
	for _, declaration := range s.other {
		if declaration.name == name {
			return declaration.value
		}
	}
	return ""
}

func (s *wrapperStyles) cssText() string {
	declarations := make([]cssDeclaration, 0, 8+len(s.other))
	if s.allBordersEqual() {
		declarations = append(declarations, cssDeclaration{name: "border", value: s.border[0].cssValue()})
	} else {
		for i, border := range s.border {
			if border.set {
				declarations = append(declarations, cssDeclaration{name: "border-" + sideNames[i], value: border.cssValue()})
			}
		}
	}
	if s.allPaddingEqual() {
		declarations = append(declarations, cssDeclaration{name: "padding", value: s.padding[0]})
	} else {
		for i, set := range s.padSet {
			if set {
				declarations = append(declarations, cssDeclaration{name: "padding-" + sideNames[i], value: s.padding[i]})
			}
		}
	}
	declarations = append(declarations, s.other...)
	var css strings.Builder
	for _, declaration := range declarations {
		if declaration.value == "" {
			continue
		}
		css.WriteString(declaration.name)
		css.WriteString(": ")
		css.WriteString(declaration.value)
		css.WriteByte(';')
		css.WriteByte(' ')
	}
	return strings.TrimSuffix(css.String(), " ")
}

func (s *wrapperStyles) allBordersEqual() bool {
	if !s.border[0].set {
		return false
	}
	for i := 1; i < len(s.border); i++ {
		if s.border[i] != s.border[0] {
			return false
		}
	}
	return true
}

func (s *wrapperStyles) allPaddingEqual() bool {
	if !s.padSet[0] {
		return false
	}
	for i := 1; i < len(s.padding); i++ {
		if !s.padSet[i] || s.padding[i] != s.padding[0] {
			return false
		}
	}
	return true
}

func (b borderStyle) cssValue() string {
	if b.raw != "" {
		return b.raw
	}
	parts := make([]string, 0, 3)
	for _, value := range []string{b.width, b.style, b.color} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " ")
}

func (w *wrapper) styleLength(value string, scale float64) float64 {
	if value == "" {
		return 0
	}
	// CommonWrapper uses one em as the relative size for percentages and
	// retains signed padding in both the outer box and content offset.
	return layout.Length2Em(value, 1, scale, w.renderer.pxPerEm)
}

func (w *wrapper) styledOuterBBox() *layout.BBox { return w.styledOuterBBoxWithSave(true) }

func (w *wrapper) styledOuterBBoxWithSave(save bool) *layout.BBox {
	bbox := w.getBBoxWithSave(save)
	if w.styles == nil {
		return bbox
	}
	outer := bbox.Clone()
	// Preserve BBox.StyleAdjust's sequence: all border widths, then padding.
	// Combining terms first changes rounding in serialized dimensions.
	outer.H += w.styleLength(w.styles.border[0].width, outer.RScale)
	outer.W += w.styleLength(w.styles.border[1].width, outer.RScale)
	outer.D += w.styleLength(w.styles.border[2].width, outer.RScale)
	outer.W += w.styleLength(w.styles.border[3].width, outer.RScale)
	outer.H += w.styleLength(w.styles.padding[0], outer.RScale)
	outer.W += w.styleLength(w.styles.padding[1], outer.RScale)
	outer.D += w.styleLength(w.styles.padding[2], outer.RScale)
	outer.W += w.styleLength(w.styles.padding[3], outer.RScale)
	return outer
}

func (w *wrapper) handleStyles(element *Element) {
	if w.styles == nil {
		return
	}
	if css := w.styles.cssText(); css != "" {
		element.SetAttr("style", css)
	}
	w.dx += w.styleLength(w.styles.border[3].width, w.bbox.RScale)
	w.dx += w.styleLength(w.styles.padding[3], w.bbox.RScale)
}

func (w *wrapper) handleBorder(element *Element) {
	if w.styles == nil {
		return
	}
	var widths [4]float64
	for i, border := range w.styles.border {
		if border.width != "" {
			// SVGWrapper clamps the drawn border independently of box sizing.
			// Its painter uses every authored width; only dotted/dashed styles
			// select a special path, while other styles use a solid polygon.
			widths[i] = math.Max(0, w.styleLength(border.width, w.bbox.RScale))
		}
	}
	bbox := w.outerBBox()
	h, d, width := bbox.H+borderFuzz, bbox.D+borderFuzz, bbox.W+borderFuzz
	outerRT, outerLT := [2]float64{width, h}, [2]float64{-borderFuzz, h}
	outerRB, outerLB := [2]float64{width, -d}, [2]float64{-borderFuzz, -d}
	innerRT := [2]float64{width - widths[1], h - widths[0]}
	innerLT := [2]float64{-borderFuzz + widths[3], h - widths[0]}
	innerRB := [2]float64{width - widths[1], -d + widths[2]}
	innerLB := [2]float64{-borderFuzz + widths[3], -d + widths[2]}
	paths := [4][4][2]float64{
		{outerLT, outerRT, innerRT, innerLT},
		{outerRB, outerRT, innerRT, innerRB},
		{outerLB, outerRB, innerRB, innerLB},
		{outerLB, outerLT, innerLT, innerLB},
	}
	var original Node
	if len(element.Children) != 0 {
		original = element.Children[0]
	}
	for i, thickness := range widths {
		if thickness == 0 {
			continue
		}
		border := w.styles.border[i]
		color := border.color
		if color == "" {
			color = "currentColor"
		}
		if border.style == "dashed" || border.style == "dotted" {
			w.addBrokenBorder(element, paths[i], color, border.style, thickness, i)
		} else {
			polygon := NewElement("polygon").
				SetAttr("points", w.borderPoints(paths[i])).
				SetAttr("stroke", "none").
				SetAttr("fill", color)
			insertBeforeNode(element, original, polygon)
		}
	}
}

func (w *wrapper) borderPoints(path [4][2]float64) string {
	points := make([]string, len(path))
	for i, point := range path {
		points[i] = fixed(point[0]-w.dx) + "," + fixed(point[1])
	}
	return strings.Join(points, " ")
}

func insertBeforeNode(parent *Element, before Node, child Node) {
	if before == nil {
		parent.Append(child)
		return
	}
	for index, candidate := range parent.Children {
		if candidate != before {
			continue
		}
		parent.Children = append(parent.Children, nil)
		copy(parent.Children[index+1:], parent.Children[index:])
		parent.Children[index] = child
		return
	}
	parent.Append(child)
}

func (w *wrapper) addBrokenBorder(element *Element, path [4][2]float64, color, style string, thickness float64, side int) {
	half := thickness / 2
	offsets := [4][4]float64{{half, -half, -half, -half}, {-half, half, -half, -half}, {half, half, -half, half}, {half, half, half, -half}}
	offset := offsets[side]
	x1 := path[0][0] + offset[0] - w.dx
	y1 := path[0][1] + offset[1]
	x2 := path[1][0] + offset[2] - w.dx
	y2 := path[1][1] + offset[3]
	length := math.Abs(x2 - x1)
	if side%2 == 1 {
		length = math.Abs(y2 - y1)
	}
	dotted := style == "dotted"
	count := math.Ceil((length - thickness) / (4 * thickness))
	if dotted {
		count = math.Ceil(length / (2 * thickness))
	}
	if count < 1 {
		count = 1
	}
	dash := ""
	linecap := "square"
	if dotted {
		linecap = "round"
		dash = "1 " + fixed(length/count-.002)
	} else {
		unit := length / (4*count + 1)
		dash = fixed(unit) + " " + fixed(3*unit)
	}
	line := NewElement("line").
		SetAttr("x1", fixed(x1)).
		SetAttr("y1", fixed(y1)).
		SetAttr("x2", fixed(x2)).
		SetAttr("y2", fixed(y2)).
		SetAttr("stroke-width", fixed(thickness)).
		SetAttr("stroke", color).
		SetAttr("stroke-linecap", linecap).
		SetAttr("stroke-dasharray", dash)
	var first Node
	if len(element.Children) != 0 {
		first = element.Children[0]
	}
	insertBeforeNode(element, first, line)
}
