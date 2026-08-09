// Copyright (c) 2018-2022 The MathJax Consortium
// Color portions Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go declarative translation of the MathJax 3.2.2 base, AMS, mathtools,
// AMS-CD, braket, cases, physics, empheq, newcommand, color, enclose,
// cancel, gensymb, and mhchem Configuration.ts files.

package tex

var mjSourcePackages = []mjSourcePackage{
	{
		Name:   "base",
		Source: "ts/input/tex/base/BaseConfiguration.ts",
		Line:   146,
		Handlers: []mjSourceHandler{
			{Kind: "character", Maps: []string{"command", "special", "letter", "digit"}},
			{Kind: "delimiter", Maps: []string{"delimiter"}},
			{Kind: "macro", Maps: []string{"delimiter", "macros", "mathchar0mi", "mathchar0mo", "mathchar7"}},
			{Kind: "environment", Maps: []string{"environment"}},
		},
		Fallbacks: []mjSourceReference{
			{Name: "character", Implementation: "Other"},
			{Name: "macro", Implementation: "csUndefined"},
			{Name: "environment", Implementation: "envUndefined"},
		},
		Items: []mjSourceReference{
			{Name: "start", Implementation: "StartItem"},
			{Name: "stop", Implementation: "StopItem"},
			{Name: "open", Implementation: "OpenItem"},
			{Name: "close", Implementation: "CloseItem"},
			{Name: "prime", Implementation: "PrimeItem"},
			{Name: "subsup", Implementation: "SubsupItem"},
			{Name: "over", Implementation: "OverItem"},
			{Name: "left", Implementation: "LeftItem"},
			{Name: "middle", Implementation: "Middle"},
			{Name: "right", Implementation: "RightItem"},
			{Name: "begin", Implementation: "BeginItem"},
			{Name: "end", Implementation: "EndItem"},
			{Name: "style", Implementation: "StyleItem"},
			{Name: "position", Implementation: "PositionItem"},
			{Name: "cell", Implementation: "CellItem"},
			{Name: "mml", Implementation: "MmlItem"},
			{Name: "fn", Implementation: "FnItem"},
			{Name: "not", Implementation: "NotItem"},
			{Name: "nonscript", Implementation: "NonscriptItem"},
			{Name: "dots", Implementation: "DotsItem"},
			{Name: "array", Implementation: "ArrayItem"},
			{Name: "eqnarray", Implementation: "EqnArrayItem"},
			{Name: "equation", Implementation: "EquationItem"},
		},
		Options: mjSourceObject{
			{Name: "maxMacros", Value: 1000},
			{Name: "baseURL", Value: mjSourceExpression("(typeof(document) === 'undefined' || document.getElementsByTagName('base').length === 0) ? '' : String(document.location).replace(/#.*$/, '')")},
		},
		Tags:           []mjSourceReference{{Name: "base", Implementation: "BaseTags"}},
		Postprocessors: []mjSourcePostprocessor{{Name: "filterNonscript", Priority: -4}},
	},
	{
		Name:   "ams",
		Source: "ts/input/tex/ams/AmsConfiguration.ts",
		Line:   51,
		Handlers: []mjSourceHandler{
			{Kind: "character", Maps: []string{"AMSmath-operatorLetter"}},
			{Kind: "delimiter", Maps: []string{"AMSsymbols-delimiter", "AMSmath-delimiter"}},
			{Kind: "macro", Maps: []string{
				"AMSsymbols-mathchar0mi", "AMSsymbols-mathchar0mo",
				"AMSsymbols-delimiter", "AMSsymbols-macros",
				"AMSmath-mathchar0mo", "AMSmath-macros", "AMSmath-delimiter",
			}},
			{Kind: "environment", Maps: []string{"AMSmath-environment"}},
		},
		Items: []mjSourceReference{
			{Name: "multline", Implementation: "MultlineItem"},
			{Name: "flalign", Implementation: "FlalignItem"},
		},
		Tags: []mjSourceReference{{Name: "ams", Implementation: "AmsTags"}},
		Options: mjSourceObject{
			{Name: "multlineWidth", Value: ""},
			{Name: "ams", Value: mjSourceObject{
				{Name: "multlineWidth", Value: "100%"},
				{Name: "multlineIndent", Value: "1em"},
			}},
		},
		DynamicMapNames: []string{"ams-declare-ops"},
		DynamicHandlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"ams-declare-ops"}}},
		DynamicPriority: -1,
		Hooks: []mjSourceHook{
			{Kind: "init", Implementation: "init"},
			{Kind: "config", Implementation: "config"},
		},
	},
	{
		Name:   "mathtools",
		Source: "ts/input/tex/mathtools/MathtoolsConfiguration.ts",
		Line:   94,
		Handlers: []mjSourceHandler{
			{Kind: "macro", Maps: []string{"mathtools-macros", "mathtools-delimiters"}},
			{Kind: "environment", Maps: []string{"mathtools-environments"}},
			{Kind: "delimiter", Maps: []string{"mathtools-delimiters"}},
			{Kind: "character", Maps: []string{"mathtools-characters"}},
		},
		Items: []mjSourceReference{{Name: "multlined", Implementation: "MultlinedItem"}},
		Options: mjSourceObject{{Name: "mathtools", Value: mjSourceObject{
			{Name: "multlinegap", Value: "1em"},
			{Name: "multlined-pos", Value: "c"},
			{Name: "firstline-afterskip", Value: ""},
			{Name: "lastline-preskip", Value: ""},
			{Name: "smallmatrix-align", Value: "c"},
			{Name: "shortvdotsadjustabove", Value: ".2em"},
			{Name: "shortvdotsadjustbelow", Value: ".2em"},
			{Name: "centercolon", Value: false},
			{Name: "centercolon-offset", Value: ".04em"},
			{Name: "thincolon-dx", Value: "-.04em"},
			{Name: "thincolon-dw", Value: "-.08em"},
			{Name: "use-unicode", Value: false},
			{Name: "prescript-sub-format", Value: ""},
			{Name: "prescript-sup-format", Value: ""},
			{Name: "prescript-arg-format", Value: ""},
			{Name: "allow-mathtoolsset", Value: true},
			{Name: "pairedDelimiters", Value: mjSourceExpandable{Value: mjSourceObject{}}},
			{Name: "tagforms", Value: mjSourceExpandable{Value: mjSourceObject{}}},
		}}},
		DynamicMapNames: []string{"mathtools-paired-delims"},
		DynamicHandlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"mathtools-paired-delims"}}},
		DynamicPriority: -5,
		Hooks: []mjSourceHook{
			{Kind: "init", Implementation: "initMathtools"},
			{Kind: "config", Implementation: "configMathtools"},
		},
		Postprocessors: []mjSourcePostprocessor{{Name: "fixPrescripts", Priority: -6}},
	},
	{
		Name:   "amscd",
		Source: "ts/input/tex/amscd/AmsCdConfiguration.ts",
		Line:   29,
		Handlers: []mjSourceHandler{
			{Kind: "character", Maps: []string{"amscd_special"}},
			{Kind: "macro", Maps: []string{"amscd_macros"}},
			{Kind: "environment", Maps: []string{"amscd_environment"}},
		},
		Options: mjSourceObject{{Name: "amscd", Value: mjSourceObject{
			{Name: "colspace", Value: "5pt"},
			{Name: "rowspace", Value: "5pt"},
			{Name: "harrowsize", Value: "2.75em"},
			{Name: "varrowsize", Value: "1.75em"},
			{Name: "hideHorizontalLabels", Value: false},
		}}},
	},
	{
		Name:   "braket",
		Source: "ts/input/tex/braket/BraketConfiguration.ts",
		Line:   30,
		Handlers: []mjSourceHandler{
			{Kind: "character", Maps: []string{"Braket-characters"}},
			{Kind: "macro", Maps: []string{"Braket-macros"}},
		},
		Items: []mjSourceReference{{Name: "braket", Implementation: "BraketItem"}},
	},
	{
		Name:   "cases",
		Source: "ts/input/tex/cases/CasesConfiguration.ts",
		Line:   203,
		Handlers: []mjSourceHandler{
			{Kind: "environment", Maps: []string{"cases-env"}},
			{Kind: "character", Maps: []string{"cases-macros"}},
		},
		Items: []mjSourceReference{{Name: "cases-begin", Implementation: "CasesBeginItem"}},
		Tags:  []mjSourceReference{{Name: "cases", Implementation: "CasesTags"}},
	},
	{
		Name:   "physics",
		Source: "ts/input/tex/physics/PhysicsConfiguration.ts",
		Line:   30,
		Handlers: []mjSourceHandler{
			{Kind: "macro", Maps: []string{
				"Physics-automatic-bracing-macros", "Physics-vector-macros",
				"Physics-vector-mo", "Physics-vector-mi", "Physics-derivative-macros",
				"Physics-expressions-macros", "Physics-quick-quad-macros",
				"Physics-bra-ket-macros", "Physics-matrix-macros",
			}},
			{Kind: "character", Maps: []string{"Physics-characters"}},
			{Kind: "environment", Maps: []string{"Physics-aux-envs"}},
		},
		Items: []mjSourceReference{{Name: "auto open", Implementation: "AutoOpen"}},
		Options: mjSourceObject{{Name: "physics", Value: mjSourceObject{
			{Name: "italicdiff", Value: false},
			{Name: "arrowdel", Value: false},
		}}},
	},
	{
		Name:   "empheq",
		Source: "ts/input/tex/empheq/EmpheqConfiguration.ts",
		Line:   172,
		Handlers: []mjSourceHandler{
			{Kind: "macro", Maps: []string{"empheq-macros"}},
			{Kind: "environment", Maps: []string{"empheq-env"}},
		},
		Items: []mjSourceReference{{Name: "empheq-begin", Implementation: "EmpheqBeginItem"}},
	},
	{
		Name:            "newcommand",
		Source:          "ts/input/tex/newcommand/NewcommandConfiguration.ts",
		Line:            54,
		Handlers:        []mjSourceHandler{{Kind: "macro", Maps: []string{"Newcommand-macros"}}},
		Items:           []mjSourceReference{{Name: "beginEnv", Implementation: "BeginEnvItem"}},
		Options:         mjSourceObject{{Name: "maxMacros", Value: 1000}},
		DynamicMapNames: []string{"new-Delimiter", "new-Command", "new-Environment"},
		DynamicHandlers: []mjSourceHandler{
			{Kind: "character", Maps: []string{}},
			{Kind: "delimiter", Maps: []string{"new-Delimiter"}},
			{Kind: "macro", Maps: []string{"new-Delimiter", "new-Command"}},
			{Kind: "environment", Maps: []string{"new-Environment"}},
		},
		DynamicPriority: -1,
		Hooks:           []mjSourceHook{{Kind: "init", Implementation: "init"}},
	},
	{
		Name:     "color",
		Source:   "ts/input/tex/color/ColorConfiguration.ts",
		Line:     55,
		Handlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"color"}}},
		Options: mjSourceObject{{Name: "color", Value: mjSourceObject{
			{Name: "padding", Value: "5px"},
			{Name: "borderWidth", Value: "2px"},
		}}},
		Hooks: []mjSourceHook{{Kind: "config", Implementation: "config"}},
	},
	{
		Name:     "enclose",
		Source:   "ts/input/tex/enclose/EncloseConfiguration.ts",
		Line:     70,
		Handlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"enclose"}}},
	},
	{
		Name:     "cancel",
		Source:   "ts/input/tex/cancel/CancelConfiguration.ts",
		Line:     83,
		Handlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"cancel"}}},
	},
	{
		Name:     "gensymb",
		Source:   "ts/input/tex/gensymb/GensymbConfiguration.ts",
		Line:     58,
		Handlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"gensymb-symbols"}}},
	},
	{
		Name:     "mhchem",
		Source:   "ts/input/tex/mhchem/MhchemConfiguration.ts",
		Line:     93,
		Handlers: []mjSourceHandler{{Kind: "macro", Maps: []string{"mhchem"}}},
	},
}

// mjD2SelectedPackageOrder is copied from D2's MathJax setup.  The three
// dependency-only packages below complete the exact transitive closure loaded
// by MathJax's component loader for that selection.
var mjD2SelectedPackageOrder = []string{
	"base", "mathtools", "ams", "amscd", "braket", "cancel", "cases",
	"color", "gensymb", "mhchem", "physics",
}

var mjD2DependencyPackages = []string{"newcommand", "enclose", "empheq"}

// mjSourceMaps is the source-order concatenation of the package mapping files.
// Package configuration order remains independently visible in
// mjSourcePackages.
var mjSourceMaps = func() []mjSourceMap {
	maps := make([]mjSourceMap, 0,
		len(mjSourceBaseMaps)+len(mjSourceAMSMaps)+len(mjSourceMathtoolsMaps)+
			len(mjSourceAmsCdBraketCasesMaps)+len(mjSourcePhysicsMaps)+
			len(mjSourceEmpheqNewcommandMaps)+len(mjSourceSelectedExtensionMaps))
	maps = append(maps, mjSourceBaseMaps...)
	maps = append(maps, mjSourceAMSMaps...)
	maps = append(maps, mjSourceMathtoolsMaps...)
	maps = append(maps, mjSourceAmsCdBraketCasesMaps...)
	maps = append(maps, mjSourcePhysicsMaps...)
	maps = append(maps, mjSourceEmpheqNewcommandMaps...)
	maps = append(maps, mjSourceSelectedExtensionMaps...)
	return maps
}()
