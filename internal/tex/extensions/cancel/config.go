// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 CancelConfiguration.ts.

// Package cancel describes MathJax 3.2.2's TeX cancel extension.
package cancel

import (
	"github.com/d2lang/mathjax-go/internal/tex/extensions/enclose"
	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

const (
	UpDiagonalStrike   = "updiagonalstrike"
	DownDiagonalStrike = "downdiagonalstrike"
	UpDiagonalArrow    = "updiagonalarrow"
	NorthEastArrow     = "northeastarrow"

	// CancelToNotation is the menclose notation built by CancelTo.
	CancelToNotation = UpDiagonalStrike + " " + UpDiagonalArrow + " " + NorthEastArrow
)

// CancelToPadding is the fixed mpadded recipe applied to cancelto's value.
var CancelToPadding = []spec.Attribute{
	{Name: "depth", Value: "-.1em"},
	{Name: "height", Value: "+.1em"},
	{Name: "voffset", Value: ".1em"},
}

// Configuration mirrors CancelConfiguration.ts. Arguments preserves the
// fixed notation argument attached to each array-valued CommandMap entry.
var Configuration = spec.Package{
	Name:     "cancel",
	MacroMap: "cancel",
	Commands: []spec.Command{
		{Name: "cancel", Handler: "Cancel", Arguments: []string{UpDiagonalStrike}},
		{Name: "bcancel", Handler: "Cancel", Arguments: []string{DownDiagonalStrike}},
		{Name: "xcancel", Handler: "Cancel", Arguments: []string{UpDiagonalStrike + " " + DownDiagonalStrike}},
		{Name: "cancelto", Handler: "CancelTo"},
	},
	AllowedOptions: enclose.AllowedOptions,
}
