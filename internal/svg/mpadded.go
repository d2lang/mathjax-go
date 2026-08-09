// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
)

type paddedDimensions struct {
	contentH float64
	contentD float64
	contentW float64
	deltaH   float64
	deltaD   float64
	deltaW   float64
	x        float64
	y        float64
	dx       float64
}

func (w *wrapper) getPaddedDimensions() paddedDimensions {
	content := layout.ZeroBBox()
	if len(w.children) != 0 {
		content = w.children[0].getBBox()
	}
	H, D, W := content.H, content.D, content.W
	h, d, width := H, D, W
	if value := stringAttribute(w.node, "width", ""); value != "" {
		width = w.paddedDimension(value, content, "w", true)
	}
	if value := stringAttribute(w.node, "height", ""); value != "" {
		h = w.paddedDimension(value, content, "h", true)
	}
	if value := stringAttribute(w.node, "depth", ""); value != "" {
		d = w.paddedDimension(value, content, "d", true)
	}
	x, y := 0.0, 0.0
	if value := stringAttribute(w.node, "voffset", ""); value != "" {
		y = w.paddedDimension(value, content, "", false)
	}
	if value := stringAttribute(w.node, "lspace", ""); value != "" {
		x = w.paddedDimension(value, content, "", false)
	}
	dx := 0.0
	if align := stringAttribute(w.node, "data-align", ""); align != "" {
		dx = paddedAlignX(width, content, align)
	}
	return paddedDimensions{
		contentH: H, contentD: D, contentW: W,
		deltaH: h - H, deltaD: d - D, deltaW: width - W,
		x: x, y: y, dx: dx,
	}
}

func (w *wrapper) paddedDimension(length string, bbox *layout.BBox, defaultDimension string, clamp bool) float64 {
	size := 0.0
	if strings.Contains(length, "width") {
		size = bbox.W
	} else if strings.Contains(length, "height") {
		size = bbox.H
	} else if strings.Contains(length, "depth") {
		size = bbox.D
	} else {
		switch defaultDimension {
		case "w":
			size = bbox.W
		case "h":
			size = bbox.H
		case "d":
			size = bbox.D
		}
	}
	value := layout.Length2Em(length, size, bbox.Scale, w.renderer.pxPerEm)
	if defaultDimension != "" && (strings.HasPrefix(length, "+") || strings.HasPrefix(length, "-")) {
		value += size
	}
	if clamp {
		value = math.Max(0, value)
	}
	return value
}

func paddedAlignX(width float64, bbox *layout.BBox, align string) float64 {
	switch align {
	case "right":
		return width - (bbox.W+bbox.R)*bbox.RScale
	case "left":
		return bbox.L * bbox.RScale
	default:
		return (width - bbox.W*bbox.RScale) / 2
	}
}

func (w *wrapper) computePaddedBBox(bbox *layout.BBox) {
	dimensions := w.getPaddedDimensions()
	bbox.W = dimensions.contentW + dimensions.deltaW
	bbox.H = dimensions.contentH + dimensions.deltaH
	bbox.D = dimensions.contentD + dimensions.deltaD
}

func (w *wrapper) paddedToSVG(parent *Element) {
	element := w.standardSVG(parent)
	dimensions := w.getPaddedDimensions()
	align := stringAttribute(w.node, "data-align", "left")
	x := dimensions.x + dimensions.dx
	if dimensions.deltaW < 0 && align != "left" {
		if align == "center" {
			x -= dimensions.deltaW / 2
		} else {
			x -= dimensions.deltaW
		}
	}
	content := element
	if x != 0 || dimensions.y != 0 {
		content = NewElement("g")
		element.Append(content)
		w.place(x, dimensions.y, content)
	}
	w.addChildren(content)
}
