// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go declarative translation of MathJax 3.2.2 amscd/AmsCdMappings.ts, braket/BraketMappings.ts, cases/CasesConfiguration.ts.

package tex

var mjSourceAmsCdBraketCasesMaps = []mjSourceMap{
	// amscd/AmsCdMappings.ts
	// Upstream line 30.
	{
		Name:    "amscd_environment",
		Kind:    mjSourceEnvironmentMap,
		Source:  "ts/input/tex/amscd/AmsCdMappings.ts",
		Line:    30,
		Parser:  "ParseMethods.environment",
		Methods: "AmsCdMethods",
		Entries: []mjSourceEntry{
			{Name: "CD", Value: "CD"},
		},
	},
	// amscd/AmsCdMappings.ts
	// Upstream line 33.
	{
		Name:    "amscd_macros",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/amscd/AmsCdMappings.ts",
		Line:    33,
		Methods: "AmsCdMethods",
		Entries: []mjSourceEntry{
			{Name: "minCDarrowwidth", Value: "minCDarrowwidth"},
			{Name: "minCDarrowheight", Value: "minCDarrowheight"},
		},
	},
	// amscd/AmsCdMappings.ts
	// Upstream line 38.
	{
		Name:    "amscd_special",
		Kind:    mjSourceMacroMap,
		Source:  "ts/input/tex/amscd/AmsCdMappings.ts",
		Line:    38,
		Methods: "AmsCdMethods",
		Entries: []mjSourceEntry{
			{Name: "@", Value: "arrow"},
		},
	},
	// braket/BraketMappings.ts
	// Upstream line 32.
	{
		Name:    "Braket-macros",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/braket/BraketMappings.ts",
		Line:    32,
		Methods: "BraketMethods",
		Entries: []mjSourceEntry{
			{Name: "bra", Value: mjSourceList{"Macro", "{\\langle {#1} \\vert}", 1}},
			{Name: "ket", Value: mjSourceList{"Macro", "{\\vert {#1} \\rangle}", 1}},
			{Name: "braket", Value: mjSourceList{"Braket", "⟨", "⟩", false, mjPositiveInfinity}},
			{Name: "set", Value: mjSourceList{"Braket", "{", "}", false, 1}},
			{Name: "Bra", Value: mjSourceList{"Macro", "{\\left\\langle {#1} \\right\\vert}", 1}},
			{Name: "Ket", Value: mjSourceList{"Macro", "{\\left\\vert {#1} \\right\\rangle}", 1}},
			{Name: "Braket", Value: mjSourceList{"Braket", "⟨", "⟩", true, mjPositiveInfinity}},
			{Name: "Set", Value: mjSourceList{"Braket", "{", "}", true, 1}},
			{Name: "ketbra", Value: mjSourceList{"Macro", "{\\vert {#1} \\rangle\\langle {#2} \\vert}", 2}},
			{Name: "Ketbra", Value: mjSourceList{"Macro", "{\\left\\vert {#1} \\right\\rangle\\left\\langle {#2} \\right\\vert}", 2}},
			{Name: "|", Value: "Bar"},
		},
	},
	// braket/BraketMappings.ts
	// Upstream line 52.
	{
		Name:    "Braket-characters",
		Kind:    mjSourceMacroMap,
		Source:  "ts/input/tex/braket/BraketMappings.ts",
		Line:    52,
		Methods: "BraketMethods",
		Entries: []mjSourceEntry{
			{Name: "|", Value: "Bar"},
		},
	},
	// cases/CasesConfiguration.ts
	// Upstream line 188.
	{
		Name:    "cases-env",
		Kind:    mjSourceEnvironmentMap,
		Source:  "ts/input/tex/cases/CasesConfiguration.ts",
		Line:    188,
		Parser:  "EmpheqUtil.environment",
		Methods: "CasesMethods",
		Entries: []mjSourceEntry{
			{Name: "numcases", Value: mjSourceList{"NumCases", "cases"}},
			{Name: "subnumcases", Value: mjSourceList{"NumCases", "cases"}},
		},
	},
	// cases/CasesConfiguration.ts
	// Upstream line 196.
	{
		Name:    "cases-macros",
		Kind:    mjSourceMacroMap,
		Source:  "ts/input/tex/cases/CasesConfiguration.ts",
		Line:    196,
		Methods: "CasesMethods",
		Entries: []mjSourceEntry{
			{Name: "&", Value: "Entry"},
		},
	},
}
