// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
)

// This is ParseUtil.matchDimen's full-token mu branch, not a replacement for
// readDimension's existing scanner. In particular, \s here is ECMAScript
// whitespace and the decimal digits are ASCII.
const muDimensionSpace = `[\t-\r \x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]`

var muDimensionPattern = regexp.MustCompile(`^` + muDimensionSpace + `*([-+]?(?:[.,][0-9]+|[0-9]+(?:[.,][0-9]*)?))` + muDimensionSpace + `*mu` + muDimensionSpace + `*$`)

// normalizeTeXMu applies ParseUtil.muReplace/Em only to a valid full mu value
// already extracted by the TeX parser. Other units, invalid values, parser
// errors and cursor movement retain their existing behavior. Direct MathML
// lengths never pass through this conversion.
func normalizeTeXMu(value string) string {
	match := muDimensionPattern.FindStringSubmatch(value)
	if match == nil {
		return value
	}
	number, err := strconv.ParseFloat(strings.Replace(match[1], ",", ".", 1), 64)
	if err != nil && !math.IsInf(number, 0) {
		return value
	}
	em := number / 18
	if math.Abs(em) < .0006 {
		return "0em"
	}
	// Preserve the primary's exact /\.?0+$/ replacement, including its
	// behavior on scientific notation returned by toFixed for large numbers.
	return strings.TrimSuffix(strings.TrimRight(jscompat.ToFixed(em, 3), "0"), ".") + "em"
}
