// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/input/tex/base/BaseMethods.ts, base/BaseItems.ts, ams/AmsMethods.ts,
// and ams/AmsItems.ts.

package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
)

// amsTagEnvironment is the narrow environment-dispatch seam used after the
// central \begin scanner has consumed the environment name.
func (p *parser) amsTagEnvironment(environment string) (nodes []*mml.Node, handled bool, err error) {
	switch environment {
	case "equation", "equation*":
		nodes, err = p.amsEquation(environment)
	case "split", "align", "align*", "aligned", "alignedat", "gather", "gather*", "eqnarray", "eqnarray*":
		nodes, err = p.amsAlignment(environment)
	default:
		return nil, false, nil
	}
	return nodes, true, err
}

func (p *parser) amsEquation(environment string) (nodes []*mml.Node, err error) {
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, err
	}
	state := p.amsTags()
	state.start("equation", true, environment == "equation")
	ended := false
	defer func() {
		if !ended {
			state.end()
		}
	}()

	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	content, err := p.parseContinuationString(body)
	if err != nil {
		return nil, err
	}
	tag, err := state.getTag(p)
	if err != nil {
		return nil, err
	}
	state.end()
	ended = true
	if tag != nil {
		return []*mml.Node{amsEnTag(content, tag)}, nil
	}
	// EquationItem returns its inferred row to the surrounding TeX stack.  The
	// stack flattens it, which is observable when \eqref follows \end{equation}.
	return unwrapInferred(content), nil
}

func (p *parser) amsAlignment(environment string) (nodes []*mml.Node, err error) {
	verticalAlign := ""
	if environment == "aligned" {
		verticalAlign, err = p.readArrayAlignment()
		if err != nil {
			return nil, err
		}
	}
	pairCount := ""
	if environment == "alignedat" {
		verticalAlign, err = p.readArrayAlignment()
		if err != nil {
			return nil, err
		}
		pairCount, err = p.readEquationPairCount(environment)
		if err != nil {
			return nil, err
		}
	}
	taggable := environment == "eqnarray" || environment == "eqnarray*" ||
		environment == "align" || environment == "align*" ||
		environment == "gather" || environment == "gather*"
	if taggable {
		if err := p.checkEquationEnvironment(); err != nil {
			return nil, err
		}
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}

	defaultTags := environment == "align" || environment == "gather" || environment == "eqnarray"
	state := p.amsTags()
	state.start(environment, taggable, defaultTags)
	ended := false
	defer func() {
		if !ended {
			state.end()
		}
	}()

	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	tags := make([]*mml.Node, 0, len(rows))
	maximumColumns := 0
	appendRow := func(row *mml.Node) error {
		if len(row.Children) > maximumColumns {
			maximumColumns = len(row.Children)
		}
		mrows = append(mrows, row)
		tag, tagErr := state.getTag(p)
		if tagErr != nil {
			return tagErr
		}
		tags = append(tags, tag)
		state.clearTag()
		return nil
	}
	spacing := newEquationTableState(appendRow)
	for rowIndex, cells := range rows {
		if err := p.parseEquationRow(cells, rowIndex == len(rows)-1, appendRow, spacing); err != nil {
			return nil, err
		}
	}

	table := node("mtable", mrows...)
	for i, tag := range tags {
		if tag == nil {
			continue
		}
		row := table.Children[i]
		table.Children[i] = node("mlabeledtr", append([]*mml.Node{tag}, row.Children...)...)
		table.Children[i].Parent = table
	}
	finishAMSEqnArrayTable(table, environment, maximumColumns)
	if environment == "alignedat" {
		finishAlignedatTable(table, pairCount, verticalAlign, maximumColumns)
		verticalAlign = ""
	}
	setArrayAlign(table, verticalAlign)

	spacing.applySpacing(table)
	state.end()
	ended = true
	return []*mml.Node{table}, nil
}
