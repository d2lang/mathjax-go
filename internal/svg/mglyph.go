// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

type glyphParameters struct {
	src, alt              string
	width, height, valign float64
	character             string
	fontFamily            string
}

func parseMglyphIndex(value string) rune {
	value = strings.TrimSpace(value)
	sign, start := int64(1), 0
	if strings.HasPrefix(value, "+") {
		start = 1
	} else if strings.HasPrefix(value, "-") {
		sign, start = -1, 1
	}
	end := start
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	if end == start {
		return 0
	}
	parsed, err := strconv.ParseInt(value[start:end], 10, 32)
	if err != nil {
		return 0
	}
	parsed *= sign
	if parsed < 0 || parsed > utf8.MaxRune || parsed >= 0xD800 && parsed <= 0xDFFF {
		return 0
	}
	return rune(parsed)
}

func (w *wrapper) getGlyphParameters() glyphParameters {
	parameters := glyphParameters{
		src: stringAttribute(w.node, "src", ""),
		alt: stringAttribute(w.node, "alt", ""),
	}
	if parameters.src != "" {
		width := stringAttribute(w.node, "width", "auto")
		height := stringAttribute(w.node, "height", "auto")
		if width == "auto" {
			parameters.width = 1
		} else {
			parameters.width = layout.Length2Em(width, 1, w.bbox.Scale, w.renderer.pxPerEm)
		}
		if height == "auto" {
			parameters.height = 1
		} else {
			parameters.height = layout.Length2Em(height, 1, w.bbox.Scale, w.renderer.pxPerEm)
		}
		parameters.valign = layout.Length2Em(stringAttribute(w.node, "valign", "0em"), 0, w.bbox.Scale, w.renderer.pxPerEm)
		return parameters
	}
	parameters.character = string(parseMglyphIndex(stringAttribute(w.node, "index", "")))
	parameters.fontFamily = stringAttribute(w.node, "fontfamily", "")
	if parameters.fontFamily != "" {
		w.explicitFont = true
		w.fontFamily = parameters.fontFamily
	}
	return parameters
}

func (w *wrapper) glyphCharacterWrapper(parameters glyphParameters) *wrapper {
	text := mml.NewText(parameters.character)
	return w.renderer.wrap(text, w, w.scriptLevel, w.displayStyle)
}

func (w *wrapper) computeGlyphBBox(bbox *layout.BBox) {
	parameters := w.getGlyphParameters()
	if parameters.src == "" {
		bbox.UpdateFrom(w.glyphCharacterWrapper(parameters).getBBox())
		return
	}
	bbox.W = parameters.width
	bbox.H = parameters.height + parameters.valign
	bbox.D = -parameters.valign
}

func (w *wrapper) glyphToSVG(parent *Element) {
	parameters := w.getGlyphParameters()
	element := w.standardSVG(parent)
	if parameters.src == "" {
		w.glyphCharacterWrapper(parameters).toSVG(element)
		return
	}
	image := NewElement("image").
		SetAttr("width", fixed(parameters.width)).
		SetAttr("height", fixed(parameters.height)).
		SetAttr("transform", "translate(0 "+fixed(parameters.height+parameters.valign)+") matrix(1 0 0 -1 0 0)").
		SetAttr("preserveAspectRatio", "none").
		SetAttr("aria-label", parameters.alt).
		SetAttr("href", parameters.src)
	element.Append(image)
}
