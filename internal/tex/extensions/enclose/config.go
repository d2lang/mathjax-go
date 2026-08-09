// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 EncloseConfiguration.ts.

// Package enclose describes MathJax 3.2.2's TeX enclose extension.
package enclose

import "github.com/d2lang/mathjax-go/internal/tex/extensions/spec"

// AllowedOptions is ENCLOSE_OPTIONS from EncloseConfiguration.ts in source
// property order. The core key-value parser should accept only these names.
var AllowedOptions = []string{
	"data-arrowhead",
	"color",
	"mathcolor",
	"background",
	"mathbackground",
	"data-padding",
	"data-thickness",
}

// Configuration mirrors EncloseConfiguration.ts.
var Configuration = spec.Package{
	Name:     "enclose",
	MacroMap: "enclose",
	Commands: []spec.Command{
		{Name: "enclose", Handler: "Enclose"},
	},
	AllowedOptions: AllowedOptions,
}
