// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of MathJax 3.2.2 ColorConfiguration.ts.

package color

import "github.com/d2lang/mathjax-go/internal/tex/extensions/spec"

// Configuration mirrors ColorConfiguration.ts. Macro order is source order.
var Configuration = spec.Package{
	Name:     "color",
	MacroMap: "color",
	Commands: []spec.Command{
		{Name: "color", Handler: "Color"},
		{Name: "textcolor", Handler: "TextColor"},
		{Name: "definecolor", Handler: "DefineColor"},
		{Name: "colorbox", Handler: "ColorBox"},
		{Name: "fcolorbox", Handler: "FColorBox"},
	},
	Options: []spec.Option{
		{Name: "padding", Value: "5px"},
		{Name: "borderWidth", Value: "2px"},
	},
}
