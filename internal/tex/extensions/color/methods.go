// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 ColorMethods.ts.

package color

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

var (
	parseFloatPrefix = regexp.MustCompile(`^[+-]?(?:Infinity|(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)`)
	trailingUnit     = regexp.MustCompile(`[a-z]*$`)
)

// PaddingProperties ports ColorMethods.ts's padding() helper. Attribute order
// matches the source PropertyList so a core parser can apply it directly to an
// mpadded node.
func PaddingProperties(colorPadding string) []spec.Attribute {
	pad := "+" + colorPadding
	unit := paddingUnit(colorPadding)
	pad2 := 2 * parseFloat(pad)
	return []spec.Attribute{
		{Name: "width", Value: "+" + jscompat.NumberString(pad2) + unit},
		{Name: "height", Value: pad},
		{Name: "depth", Value: pad},
		{Name: "lspace", Value: colorPadding},
	}
}

func paddingUnit(value string) string {
	// JavaScript's dot does not match line terminators, so the source
	// /^.*?([a-z]*)$/ replacement leaves such a value entirely unchanged.
	if strings.ContainsAny(value, "\n\r\u2028\u2029") {
		return value
	}
	return trailingUnit.FindString(value)
}

func parseFloat(value string) float64 {
	prefix := parseFloatPrefix.FindString(value)
	if prefix == "" {
		return math.NaN()
	}
	if prefix == "+Infinity" || prefix == "Infinity" {
		return math.Inf(1)
	}
	if prefix == "-Infinity" {
		return math.Inf(-1)
	}
	n, err := strconv.ParseFloat(prefix, 64)
	if err != nil && !math.IsInf(n, 0) {
		return math.NaN()
	}
	return n
}
