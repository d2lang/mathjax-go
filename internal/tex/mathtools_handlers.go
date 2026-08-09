// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts,
// MathtoolsUtil.ts, MathtoolsItems.ts, and MathtoolsMappings.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// mathtoolsCommand is the single command-dispatch hook for the central parser.
// It must run after user macros and before the generic source-symbol and legacy
// paired-delimiter paths, since Mathtools' dynamic delimiter map has priority
// -5 and several entries have source-visible handler behavior.
func (p *parser) mathtoolsCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	if definition, ok := p.state.pairedDelimiters[name]; ok {
		nodes, err = p.mathtoolsPairedDelimiter(name, definition)
		return nodes, true, err
	}

	switch name {
	case "mathllap", "mathrlap", "mathclap", "crampedllap", "crampedrlap", "crampedclap", "clap", "textllap", "textrlap", "textclap":
		nodes, err = p.lap(name)
	case "cramped":
		nodes, err = p.cramped(name)
	case "mathmbox":
		nodes, err = p.mathMBox(name)
	case "mathmakebox":
		nodes, err = p.mathMakeBox(name)
	case "overbracket", "underbracket":
		nodes, err = p.mathtoolsUnderOverBracket(name)
	case "DeclarePairedDelimiter", "DeclarePairedDelimiters", "DeclarePairedDelimiterX", "DeclarePairedDelimitersX", "DeclarePairedDelimiterXPP", "DeclarePairedDelimitersXPP":
		err = p.mathtoolsDeclarePairedDelimiter(name)
	case "centercolon":
		nodes = []*mml.Node{p.mathtoolsCenterColon(true, true, false)}
	case "ordinarycolon":
		nodes = []*mml.Node{p.mathtoolsCenterColon(false, false, false)}
	case "MTThinColon":
		nodes = []*mml.Node{p.mathtoolsCenterColon(true, true, true)}
	case "coloneqq", "Coloneqq", "coloneq", "Coloneq", "eqqcolon", "Eqqcolon", "eqcolon", "Eqcolon", "colonapprox", "Colonapprox", "colonsim", "Colonsim", "dblcolon":
		nodes, err = p.mathtoolsRelation(name)
	case "nuparrow", "ndownarrow":
		nodes = []*mml.Node{mathtoolsNArrow(name)}
	case "splitfrac", "splitdfrac":
		nodes, err = p.mathtoolsSplitFrac(name, name == "splitdfrac")
	case "xmathstrut":
		nodes, err = p.mathtoolsXMathStrut(name)
	case "prescript":
		nodes, err = p.prescript(name)
	case "adjustlimits":
		nodes, err = p.mathtoolsAdjustLimits(name)
	case "mathtoolsset":
		err = p.mathtoolsSetOptions(name)
	case "newtagform", "renewtagform":
		err = p.mathtoolsNewTagForm(name, name == "renewtagform")
	case "usetagform":
		err = p.mathtoolsUseTagForm(name)
	case "Aboxed", "ArrowBetweenLines", "MTFlushSpaceAbove", "MTFlushSpaceBelow", "vdotswithin", "shortvdotswithin":
		err = texError("NotInAlignment", "\\%s can only be used in aligment environments", name)
	case "shoveleft", "shoveright":
		err = texError("CommandInMultlined", "\\%s can only appear within the multline or multlined environments", name)
	default:
		return nil, false, nil
	}
	return nodes, true, err
}

// mathtoolsEnvironment is the single begin-environment hook.  p.pos must be
// immediately after the parsed environment-name argument; the handler owns
// any following options/arguments and capture of its matching \end.
func (p *parser) mathtoolsEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	switch name {
	case "smallmatrix", "smallmatrix*",
		"psmallmatrix", "psmallmatrix*", "bsmallmatrix", "bsmallmatrix*",
		"Bsmallmatrix", "Bsmallmatrix*", "vsmallmatrix", "vsmallmatrix*",
		"Vsmallmatrix", "Vsmallmatrix*":
		nodes, err = p.mathtoolsSmallMatrix(name)
	case "multlined":
		nodes, err = p.mathtoolsMultlined(name)
	case "spreadlines":
		nodes, err = p.mathtoolsSpreadLines(name)
	case "cases*", "dcases*", "rcases*", "drcases*":
		nodes, err = p.mathtoolsCases(name)
	case "align", "align*", "alignat", "alignat*", "xalignat", "xalignat*", "xxalignat", "aligned", "alignedat", "gather", "gather*", "gathered", "multline", "multline*", "lgathered", "rgathered":
		if !mathtoolsEnvironmentHasSpecial(p.source[p.pos:]) {
			return nil, false, nil
		}
		nodes, err = p.mathtoolsAlignment(name)
	default:
		return nil, false, nil
	}
	return nodes, true, err
}
