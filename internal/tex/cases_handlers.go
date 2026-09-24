// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/cases/CasesConfiguration.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// casesEnvironment owns numcases and subnumcases, including the package's
// special second-column text parsing and Empheq-based left block.
func (p *parser) casesEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if name != "numcases" && name != "subnumcases" {
		return nil, false, nil
	}
	left, _, err := p.readArgument("begin{"+name+"}", false)
	if err != nil {
		return nil, true, err
	}
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, true, err
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for _, cells := range rows {
		if len(cells) > 2 {
			return nil, true, texError("ExtraCasesAlignTab", "Extra alignment tab in text for numcase environment")
		}
		mtds := make([]*mml.Node, 0, len(cells))
		for column, raw := range cells {
			if column == 1 {
				mtds = append(mtds, node("mtd", node("mstyle", node("mtext", mml.NewText(strings.TrimLeft(raw, " \t\r\n"))))))
				continue
			}
			content, err := p.parseContinuationString(strings.TrimSpace(raw))
			if err != nil {
				return nil, true, err
			}
			mtds = append(mtds, node("mtd", content))
		}
		mrows = append(mrows, node("mtr", mtds...))
	}
	table := node("mtable", mrows...)
	table.Attributes.Set("displaystyle", false)
	table.Attributes.Set("rowspacing", ".2em")
	table.Attributes.Set("columnalign", "left left")
	table.Attributes.Set("columnspacing", "1em")
	original := p.copyNode(table)
	if !p.state.augmentedPackages {
		// D2 activates cases without activating EmpheqConfiguration.  The
		// Cases implementation reaches EmpheqUtil internally, but its emitted
		// delimiter command is therefore unresolved in the configured parser.
		return nil, true, texError("UndefinedControlSequence", "Undefined control sequence \\empheqlbrace")
	}
	if err := p.empheqAddLeft(table, original, left+"\\empheqlbrace\\,"); err != nil {
		return nil, true, err
	}
	return []*mml.Node{table}, true, nil
}
