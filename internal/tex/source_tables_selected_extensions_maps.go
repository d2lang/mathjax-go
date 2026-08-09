// Copyright (c) 2018-2022 The MathJax Consortium
// Color portions Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go declarative translation of MathJax 3.2.2 color/ColorConfiguration.ts, enclose/EncloseConfiguration.ts, cancel/CancelConfiguration.ts, gensymb/GensymbConfiguration.ts, mhchem/MhchemConfiguration.ts.

package tex

var mjSourceSelectedExtensionMaps = []mjSourceMap{
	// color/ColorConfiguration.ts
	// Upstream line 34.
	{
		Name:    "color",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/color/ColorConfiguration.ts",
		Line:    34,
		Methods: "ColorMethods",
		Entries: []mjSourceEntry{
			{Name: "color", Value: "Color"},
			{Name: "textcolor", Value: "TextColor"},
			{Name: "definecolor", Value: "DefineColor"},
			{Name: "colorbox", Value: "ColorBox"},
			{Name: "fcolorbox", Value: "FColorBox"},
		},
	},
	// enclose/EncloseConfiguration.ts
	// Upstream line 67.
	{
		Name:    "enclose",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/enclose/EncloseConfiguration.ts",
		Line:    67,
		Methods: "EncloseMethods",
		Entries: []mjSourceEntry{
			{Name: "enclose", Value: "Enclose"},
		},
	},
	// cancel/CancelConfiguration.ts
	// Upstream line 74.
	{
		Name:    "cancel",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/cancel/CancelConfiguration.ts",
		Line:    74,
		Methods: "CancelMethods",
		Entries: []mjSourceEntry{
			{Name: "cancel", Value: mjSourceList{"Cancel", "updiagonalstrike"}},
			{Name: "bcancel", Value: mjSourceList{"Cancel", "downdiagonalstrike"}},
			{Name: "xcancel", Value: mjSourceList{"Cancel", "updiagonalstrike downdiagonalstrike"}},
			{Name: "cancelto", Value: "CancelTo"},
		},
	},
	// gensymb/GensymbConfiguration.ts
	// Upstream line 49.
	{
		Name:   "gensymb-symbols",
		Kind:   mjSourceCharacterMap,
		Source: "ts/input/tex/gensymb/GensymbConfiguration.ts",
		Line:   49,
		Parser: "mathcharUnit",
		Entries: []mjSourceEntry{
			{Name: "ohm", Value: "Ω"},
			{Name: "degree", Value: "°"},
			{Name: "celsius", Value: "℃"},
			{Name: "perthousand", Value: "‰"},
			{Name: "micro", Value: "µ"},
		},
	},
	// mhchem/MhchemConfiguration.ts
	// Upstream line 57.
	{
		Name:    "mhchem",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/mhchem/MhchemConfiguration.ts",
		Line:    57,
		Methods: "MhchemMethods",
		Entries: []mjSourceEntry{
			{Name: "ce", Value: mjSourceList{"Machine", "ce"}},
			{Name: "pu", Value: mjSourceList{"Machine", "pu"}},
			{Name: "longrightleftharpoons", Value: mjSourceList{"Macro", "\\stackrel{\\textstyle{-}\\!\\!{\\rightharpoonup}}{\\smash{{\\leftharpoondown}\\!\\!{-}}}"}},
			{Name: "longRightleftharpoons", Value: mjSourceList{"Macro", "\\stackrel{\\textstyle{-}\\!\\!{\\rightharpoonup}}{\\smash{\\leftharpoondown}}"}},
			{Name: "longLeftrightharpoons", Value: mjSourceList{"Macro", "\\stackrel{\\textstyle\\vphantom{{-}}{\\rightharpoonup}}{\\smash{{\\leftharpoondown}\\!\\!{-}}}"}},
			{Name: "longleftrightarrows", Value: mjSourceList{"Macro", "\\stackrel{\\longrightarrow}{\\smash{\\longleftarrow}\\Rule{0px}{.25em}{0px}}"}},
			{Name: "tripledash", Value: mjSourceList{"Macro", "\\vphantom{-}\\raise2mu{\\kern2mu\\tiny\\text{-}\\kern1mu\\text{-}\\kern1mu\\text{-}\\kern2mu}"}},
			{Name: "xleftrightarrow", Value: mjSourceList{"xArrow", 8596, 6, 6}},
			{Name: "xrightleftharpoons", Value: mjSourceList{"xArrow", 8652, 5, 7}},
			{Name: "xRightleftharpoons", Value: mjSourceList{"xArrow", 8652, 5, 7}},
			{Name: "xLeftrightharpoons", Value: mjSourceList{"xArrow", 8652, 5, 7}},
		},
	},
}
