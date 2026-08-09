// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go declarative translation of MathJax 3.2.2 empheq/EmpheqConfiguration.ts, newcommand/NewcommandMappings.ts, newcommand/NewcommandConfiguration.ts.

package tex

var mjSourceEmpheqNewcommandMaps = []mjSourceMap{
	// empheq/EmpheqConfiguration.ts
	// Upstream line 123.
	{
		Name:    "empheq-env",
		Kind:    mjSourceEnvironmentMap,
		Source:  "ts/input/tex/empheq/EmpheqConfiguration.ts",
		Line:    123,
		Parser:  "EmpheqUtil.environment",
		Methods: "EmpheqMethods",
		Entries: []mjSourceEntry{
			{Name: "empheq", Value: mjSourceList{"Empheq", "empheq"}},
		},
	},
	// empheq/EmpheqConfiguration.ts
	// Upstream line 130.
	{
		Name:    "empheq-macros",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/empheq/EmpheqConfiguration.ts",
		Line:    130,
		Methods: "EmpheqMethods",
		Entries: []mjSourceEntry{
			{Name: "empheqlbrace", Value: mjSourceList{"EmpheqMO", "{"}},
			{Name: "empheqrbrace", Value: mjSourceList{"EmpheqMO", "}"}},
			{Name: "empheqlbrack", Value: mjSourceList{"EmpheqMO", "["}},
			{Name: "empheqrbrack", Value: mjSourceList{"EmpheqMO", "]"}},
			{Name: "empheqlangle", Value: mjSourceList{"EmpheqMO", "⟨"}},
			{Name: "empheqrangle", Value: mjSourceList{"EmpheqMO", "⟩"}},
			{Name: "empheqlparen", Value: mjSourceList{"EmpheqMO", "("}},
			{Name: "empheqrparen", Value: mjSourceList{"EmpheqMO", ")"}},
			{Name: "empheqlvert", Value: mjSourceList{"EmpheqMO", "|"}},
			{Name: "empheqrvert", Value: mjSourceList{"EmpheqMO", "|"}},
			{Name: "empheqlVert", Value: mjSourceList{"EmpheqMO", "‖"}},
			{Name: "empheqrVert", Value: mjSourceList{"EmpheqMO", "‖"}},
			{Name: "empheqlfloor", Value: mjSourceList{"EmpheqMO", "⌊"}},
			{Name: "empheqrfloor", Value: mjSourceList{"EmpheqMO", "⌋"}},
			{Name: "empheqlceil", Value: mjSourceList{"EmpheqMO", "⌈"}},
			{Name: "empheqrceil", Value: mjSourceList{"EmpheqMO", "⌉"}},
			{Name: "empheqbiglbrace", Value: mjSourceList{"EmpheqMO", "{"}},
			{Name: "empheqbigrbrace", Value: mjSourceList{"EmpheqMO", "}"}},
			{Name: "empheqbiglbrack", Value: mjSourceList{"EmpheqMO", "["}},
			{Name: "empheqbigrbrack", Value: mjSourceList{"EmpheqMO", "]"}},
			{Name: "empheqbiglangle", Value: mjSourceList{"EmpheqMO", "⟨"}},
			{Name: "empheqbigrangle", Value: mjSourceList{"EmpheqMO", "⟩"}},
			{Name: "empheqbiglparen", Value: mjSourceList{"EmpheqMO", "("}},
			{Name: "empheqbigrparen", Value: mjSourceList{"EmpheqMO", ")"}},
			{Name: "empheqbiglvert", Value: mjSourceList{"EmpheqMO", "|"}},
			{Name: "empheqbigrvert", Value: mjSourceList{"EmpheqMO", "|"}},
			{Name: "empheqbiglVert", Value: mjSourceList{"EmpheqMO", "‖"}},
			{Name: "empheqbigrVert", Value: mjSourceList{"EmpheqMO", "‖"}},
			{Name: "empheqbiglfloor", Value: mjSourceList{"EmpheqMO", "⌊"}},
			{Name: "empheqbigrfloor", Value: mjSourceList{"EmpheqMO", "⌋"}},
			{Name: "empheqbiglceil", Value: mjSourceList{"EmpheqMO", "⌈"}},
			{Name: "empheqbigrceil", Value: mjSourceList{"EmpheqMO", "⌉"}},
			{Name: "empheql", Value: "EmpheqDelim"},
			{Name: "empheqr", Value: "EmpheqDelim"},
			{Name: "empheqbigl", Value: "EmpheqDelim"},
			{Name: "empheqbigr", Value: "EmpheqDelim"},
		},
	},
	// newcommand/NewcommandMappings.ts
	// Upstream line 32.
	{
		Name:    "Newcommand-macros",
		Kind:    mjSourceCommandMap,
		Source:  "ts/input/tex/newcommand/NewcommandMappings.ts",
		Line:    32,
		Methods: "NewcommandMethods",
		Entries: []mjSourceEntry{
			{Name: "newcommand", Value: "NewCommand"},
			{Name: "renewcommand", Value: "NewCommand"},
			{Name: "newenvironment", Value: "NewEnvironment"},
			{Name: "renewenvironment", Value: "NewEnvironment"},
			{Name: "def", Value: "MacroDef"},
			{Name: "let", Value: "Let"},
		},
	},
	// newcommand/NewcommandConfiguration.ts
	// Upstream line 38.
	{
		Name:     "new-Delimiter",
		Kind:     mjSourceDelimiterMap,
		Source:   "ts/input/tex/newcommand/NewcommandConfiguration.ts",
		Line:     38,
		Parser:   "ParseMethods.delimiter",
		Dynamic:  true,
		Priority: -1,
		Entries:  []mjSourceEntry{},
	},
	// newcommand/NewcommandConfiguration.ts
	// Upstream line 40.
	{
		Name:     "new-Command",
		Kind:     mjSourceCommandMap,
		Source:   "ts/input/tex/newcommand/NewcommandConfiguration.ts",
		Line:     40,
		Dynamic:  true,
		Priority: -1,
		Entries:  []mjSourceEntry{},
	},
	// newcommand/NewcommandConfiguration.ts
	// Upstream line 41.
	{
		Name:     "new-Environment",
		Kind:     mjSourceEnvironmentMap,
		Source:   "ts/input/tex/newcommand/NewcommandConfiguration.ts",
		Line:     41,
		Parser:   "ParseMethods.environment",
		Dynamic:  true,
		Priority: -1,
		Entries:  []mjSourceEntry{},
	},
}
