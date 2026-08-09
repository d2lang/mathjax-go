// Copyright (c) 2021-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 GensymbConfiguration.ts.

// Package gensymb describes MathJax 3.2.2's TeX gensymb extension.
package gensymb

import "github.com/d2lang/mathjax-go/internal/tex/extensions/spec"

var unitAttributes = []spec.Attribute{
	{Name: "mathvariant", Value: "normal"},
	{Name: "class", Value: "MathML-Unit"},
}

// Configuration mirrors GensymbConfiguration.ts and its CharacterMap.
// Symbols retain source order and use the exact Unicode scalar values.
var Configuration = spec.Package{
	Name:     "gensymb",
	MacroMap: "gensymb-symbols",
	Symbols: []spec.Symbol{
		{Name: "ohm", Character: "\u2126", TokenKind: "mi", Attributes: unitAttributes},
		{Name: "degree", Character: "\u00B0", TokenKind: "mi", Attributes: unitAttributes},
		{Name: "celsius", Character: "\u2103", TokenKind: "mi", Attributes: unitAttributes},
		{Name: "perthousand", Character: "\u2030", TokenKind: "mi", Attributes: unitAttributes},
		{Name: "micro", Character: "\u00B5", TokenKind: "mi", Attributes: unitAttributes},
	},
}
