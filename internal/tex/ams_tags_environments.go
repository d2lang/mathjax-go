// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/input/tex/base/BaseMethods.ts, base/BaseItems.ts, ams/AmsMethods.ts,
// and ams/AmsItems.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// amsTagEnvironment is the narrow environment-dispatch seam used after the
// central \begin scanner has consumed the environment name.
func (p *parser) amsTagEnvironment(environment string) (nodes []*mml.Node, handled bool, err error) {
	switch environment {
	case "equation", "equation*":
		nodes, err = p.amsEquation(environment)
	case "split", "align", "align*", "aligned":
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
		verticalAlign, _, err = p.readBrackets(nil)
		if err != nil {
			return nil, err
		}
	}
	taggable := environment == "align" || environment == "align*"
	if taggable {
		if err := p.checkEquationEnvironment(); err != nil {
			return nil, err
		}
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}

	defaultTags := environment == "align"
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
	for _, cells := range rows {
		if len(cells) == 1 && strings.TrimSpace(cells[0]) == "" && len(rows) > 1 {
			continue
		}
		mtds := make([]*mml.Node, 0, len(cells))
		for _, cell := range cells {
			content, parseErr := p.parseContinuationString(strings.TrimSpace(cell))
			if parseErr != nil {
				return nil, parseErr
			}
			mtds = append(mtds, node("mtd", content))
		}
		if len(mtds) > maximumColumns {
			maximumColumns = len(mtds)
		}
		mrows = append(mrows, node("mtr", mtds...))
		tag, tagErr := state.getTag(p)
		if tagErr != nil {
			return nil, tagErr
		}
		tags = append(tags, tag)
		state.clearTag()
	}

	table := node("mtable", mrows...)
	prefixRelationColumns(table)
	for i, tag := range tags {
		if tag == nil {
			continue
		}
		row := table.Children[i]
		table.Children[i] = node("mlabeledtr", append([]*mml.Node{tag}, row.Children...)...)
		table.Children[i].Parent = table
	}
	finishAMSEqnArrayTable(table, environment, maximumColumns)
	switch strings.TrimSpace(verticalAlign) {
	case "t":
		table.Attributes.Set("align", "baseline 1")
	case "b":
		table.Attributes.Set("align", "baseline -1")
	case "c":
		table.Attributes.Set("align", "axis")
	case "":
	default:
		table.Attributes.Set("align", strings.TrimSpace(verticalAlign))
	}

	state.end()
	ended = true
	return []*mml.Node{table}, nil
}
