// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/ams/AmsMethods.ts (AlignAt, AmsEqnArray),
// base/BaseMethods.ts (EqnArray), and ParseUtil.ts (setArrayAlign).

package tex

import (
	"errors"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) readArrayAlignment() (string, error) {
	align, _, err := p.readBrackets(nil)
	var failure *Error
	if errors.As(err, &failure) && failure.ID == "MissingCloseBracket" {
		return "", texError("MissingCloseBracket", "Could not find closing ']' for argument to \\begin")
	}
	return align, err
}

func finishAlignedatTable(table *mml.Node, count, verticalAlign string, maximum int) {
	// AlignAt's positive counts supply repeated rl/zero-gap patterns, not a
	// maximum number of authored columns. EqnArray extends those patterns to
	// the actual row width. Empty and all-zero counts instead retain the
	// empty alignment and BaseMethods.EqnArray's fallback spacing.
	align, spacing := "right left", "0em 0em"
	if strings.TrimLeft(count, "0") == "" {
		align, spacing = "", "1em"
	}
	table.Attributes.Set("columnalign", repeatAMSEqnArrayDefinition(align, maximum))
	table.Attributes.Set("columnspacing", repeatAMSEqnArrayDefinition(spacing, maximum-1))

	setArrayAlign(table, verticalAlign)
}

func setArrayAlign(table *mml.Node, verticalAlign string) {
	// ParseUtil.trimSpaces uses JavaScript whitespace and preserves the last
	// control-space before setArrayAlign interprets t/b/c.
	align := strings.TrimFunc(verticalAlign, internalTextSpace)
	if strings.HasSuffix(align, `\`) && strings.HasSuffix(verticalAlign, " ") {
		align += " "
	}
	switch align {
	case "t":
		align = "baseline 1"
	case "b":
		align = "baseline -1"
	case "c":
		align = "axis"
	}
	if align != "" {
		table.Attributes.Set("align", align)
	}
}
