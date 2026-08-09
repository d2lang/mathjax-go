// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/jscompat"
)

const (
	verticalFuzz   = .1
	horizontalFuzz = .1
)

func (w *wrapper) stretchPart(index int) (rune, font.Variant, font.Metrics, bool) {
	if index < 0 || index >= len(w.stretch.Stretch) || w.stretch.Stretch[index] == 0 {
		return 0, "", font.Metrics{}, false
	}
	variant, ok := w.stretch.StretchVariant(index)
	if !ok {
		return 0, "", font.Metrics{}, false
	}
	codepoint := w.stretch.Stretch[index]
	glyph, ok := font.Lookup(variant, codepoint)
	if !ok {
		return 0, "", font.Metrics{}, false
	}
	return codepoint, variant, glyph.Metrics, true
}

func (w *wrapper) addStretchGlyph(element *Element, index int, x, y float64) float64 {
	codepoint, variant, metrics, ok := w.stretchPart(index)
	if !ok {
		return 0
	}
	w.placeChar(codepoint, x, y, element, variant)
	return metrics.Width
}

func (w *wrapper) stretchVertical(element *Element) {
	bbox := w.getBBox()
	height, depth, width := bbox.H, bbox.D, bbox.W
	top := w.addTop(element, height, width)
	bottom := w.addBottom(element, depth, width)
	if len(w.stretch.Stretch) == 4 {
		middleHeight, middleDepth := w.addMiddleVertical(element, width)
		w.addExtenderVertical(element, height, 0, top, middleHeight, width)
		w.addExtenderVertical(element, 0, depth, middleDepth, bottom, width)
	} else {
		w.addExtenderVertical(element, height, depth, top, bottom, width)
	}
}

func (w *wrapper) addTop(element *Element, height, width float64) float64 {
	_, _, metrics, ok := w.stretchPart(0)
	if !ok {
		return 0
	}
	w.addStretchGlyph(element, 0, (width-metrics.Width)/2, height-metrics.Height)
	return metrics.Height + metrics.Depth
}

func (w *wrapper) addBottom(element *Element, depth, width float64) float64 {
	_, _, metrics, ok := w.stretchPart(2)
	if !ok {
		return 0
	}
	w.addStretchGlyph(element, 2, (width-metrics.Width)/2, metrics.Depth-depth)
	return metrics.Height + metrics.Depth
}

func (w *wrapper) addMiddleVertical(element *Element, width float64) (float64, float64) {
	_, _, metrics, ok := w.stretchPart(3)
	if !ok {
		return 0, 0
	}
	y := (metrics.Depth-metrics.Height)/2 + w.renderer.params.Axis
	w.addStretchGlyph(element, 3, (width-metrics.Width)/2, y)
	return metrics.Height + y, metrics.Depth - y
}

func (w *wrapper) addExtenderVertical(element *Element, height, depth, top, bottom, width float64) {
	codepoint, variant, metrics, ok := w.stretchPart(1)
	if !ok {
		return
	}
	top = math.Max(0, top-verticalFuzz)
	bottom = math.Max(0, bottom-verticalFuzz)
	extent := height + depth - top - bottom
	if extent <= 0 || metrics.Height+metrics.Depth == 0 {
		return
	}
	scale := 1.5 * extent / (metrics.Height + metrics.Depth)
	y := (scale*(metrics.Height-metrics.Depth) - extent) / 2
	nested := NewElement("svg").
		SetAttr("width", fixed(metrics.Width)).
		SetAttr("height", fixed(extent)).
		SetAttr("y", fixed(bottom-depth)).
		SetAttr("x", fixed((width-metrics.Width)/2)).
		SetAttr("viewBox", strings.Join([]string{fixed(y), fixed(metrics.Width), fixed(extent)}, " "))
	nested.Attributes[len(nested.Attributes)-1].Value = "0 " + nested.Attributes[len(nested.Attributes)-1].Value
	w.placeChar(codepoint, 0, 0, nested, variant)
	if glyph := lastElement(nested); glyph != nil {
		glyph.SetAttr("transform", "scale(1,"+jscompat.Fixed(scale, 3)+")")
	}
	element.Append(nested)
}

func (w *wrapper) stretchHorizontal(element *Element) {
	width := w.getBBox().W
	left := w.addLeft(element)
	right := w.addRight(element, width)
	if len(w.stretch.Stretch) == 4 {
		middleLeft, middleRight := w.addMiddleHorizontal(element, width)
		half := width / 2
		w.addExtenderHorizontal(element, half, left, half-middleLeft, 0)
		w.addExtenderHorizontal(element, half, middleRight-half, right, half)
	} else {
		w.addExtenderHorizontal(element, width, left, right, 0)
	}
}

func (w *wrapper) addLeft(element *Element) float64 {
	return w.addStretchGlyph(element, 0, 0, 0)
}

func (w *wrapper) addRight(element *Element, width float64) float64 {
	_, _, metrics, ok := w.stretchPart(2)
	if !ok {
		return 0
	}
	w.addStretchGlyph(element, 2, width-metrics.Width, 0)
	return metrics.Width
}

func (w *wrapper) addMiddleHorizontal(element *Element, width float64) (float64, float64) {
	_, _, metrics, ok := w.stretchPart(3)
	if !ok {
		return 0, 0
	}
	x := (width - metrics.Width) / 2
	w.addStretchGlyph(element, 3, x, 0)
	return x, x + metrics.Width
}

func (w *wrapper) addExtenderHorizontal(element *Element, width, left, right, x float64) {
	codepoint, variant, metrics, ok := w.stretchPart(1)
	if !ok {
		return
	}
	right = math.Max(0, right-horizontalFuzz)
	left = math.Max(0, left-horizontalFuzz)
	extent := width - left - right
	if extent <= 0 || metrics.Width == 0 {
		return
	}
	height := metrics.Height + metrics.Depth + 2*verticalFuzz
	scale := 1.5 * extent / metrics.Width
	bottom := -(metrics.Depth + verticalFuzz)
	nested := NewElement("svg").
		SetAttr("width", fixed(extent)).
		SetAttr("height", fixed(height)).
		SetAttr("x", fixed(x+left)).
		SetAttr("y", fixed(bottom)).
		SetAttr("viewBox", strings.Join([]string{
			fixed((scale*metrics.Width - extent) / 2), fixed(bottom), fixed(extent), fixed(height),
		}, " "))
	w.placeChar(codepoint, 0, 0, nested, variant)
	if glyph := lastElement(nested); glyph != nil {
		glyph.SetAttr("transform", "scale("+jscompat.Fixed(scale, 3)+",1)")
	}
	element.Append(nested)
}

func lastElement(element *Element) *Element {
	if element == nil || len(element.Children) == 0 {
		return nil
	}
	child, _ := element.Children[len(element.Children)-1].(*Element)
	return child
}
