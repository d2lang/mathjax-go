// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package layout

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
)

var absoluteUnits = map[string]float64{
	"px": 1,
	"in": 96,
	"cm": 96 / 2.54,
	"mm": 96 / 25.4,
}

var relativeUnits = map[string]float64{
	"em": 1,
	"ex": .431,
	"pt": .1,
	"pc": 1.2,
	"mu": 1.0 / 18,
}

// MathSpace contains MathJax's named math-space and size values.
var MathSpace = map[string]float64{
	"veryverythinmathspace":          1.0 / 18,
	"verythinmathspace":              2.0 / 18,
	"thinmathspace":                  3.0 / 18,
	"mediummathspace":                4.0 / 18,
	"thickmathspace":                 5.0 / 18,
	"verythickmathspace":             6.0 / 18,
	"veryverythickmathspace":         7.0 / 18,
	"negativeveryverythinmathspace":  -1.0 / 18,
	"negativeverythinmathspace":      -2.0 / 18,
	"negativethinmathspace":          -3.0 / 18,
	"negativemediummathspace":        -4.0 / 18,
	"negativethickmathspace":         -5.0 / 18,
	"negativeverythickmathspace":     -6.0 / 18,
	"negativeveryverythickmathspace": -7.0 / 18,
	"thin":                           .04,
	"medium":                         .06,
	"thick":                          .1,
	"normal":                         1,
	"big":                            2,
	"small":                          1 / math.Sqrt2,
	"infinity":                       BigDimen,
}

var lengthPattern = regexp.MustCompile(`^\s*([-+]?(?:\.\d+|\d+(?:\.\d*)?))?(pt|em|ex|mu|px|pc|in|mm|cm|%)?`)

// Length2Em converts a MathML dimension into ems with MathJax 3.2.2's
// permissive prefix parsing.
func Length2Em(length any, size, scale, em float64) float64 {
	var value string
	switch x := length.(type) {
	case nil:
		return size
	case string:
		value = x
	case float64:
		value = jscompat.NumberString(x)
	case float32:
		value = jscompat.NumberString(float64(x))
	case int:
		value = strconv.Itoa(x)
	case int64:
		value = strconv.FormatInt(x, 10)
	default:
		value = fmt.Sprint(x)
	}
	if value == "" {
		return size
	}
	if space, ok := MathSpace[value]; ok {
		return space
	}
	match := lengthPattern.FindStringSubmatch(value)
	if match == nil {
		return size
	}
	m := 1.0
	if match[1] != "" {
		parsed, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return size
		}
		m = parsed
	}
	unit := match[2]
	if factor, ok := absoluteUnits[unit]; ok {
		return m * factor / em / scale
	}
	if factor, ok := relativeUnits[unit]; ok {
		return m * factor
	}
	if unit == "%" {
		return m / 100 * size
	}
	return m * size
}

// Percent formats a multiplier as a CSS percentage.
func Percent(m float64) string { return trimFixed(100*m, 1) + "%" }

// Em formats a dimension as ems.
func Em(m float64) string {
	if math.Abs(m) < .001 {
		return "0"
	}
	return trimFixed(m, 3) + "em"
}

// EmRounded rounds a dimension to CSS pixel boundaries before formatting.
func EmRounded(m, em float64) string {
	m = (jscompat.Round(m*em) + .05) / em
	if math.Abs(m) < .001 {
		return "0em"
	}
	return trimFixed(m, 3) + "em"
}

// PX formats an em dimension as pixels with an optional minimum.
func PX(m, minimum, em float64) string {
	m *= em
	if minimum != 0 && m < minimum {
		m = minimum
	}
	if math.Abs(m) < .1 {
		return "0"
	}
	return trimFixed(m, 1) + "px"
}

func trimFixed(value float64, digits int) string {
	return strings.TrimRight(strings.TrimRight(jscompat.ToFixed(value, digits), "0"), ".")
}
