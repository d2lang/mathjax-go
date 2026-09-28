// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts (ArrowBetweenLines),
// MathtoolsUtil.ts (checkAlignment), and base/BaseItems.ts (EqnArrayItem).
package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Only EqnArrayItem owns this callback. In particular, FlalignItem has a
// different kind even though its implementation extends EqnArrayItem.
type equationRowState struct {
	columns int
	endRow  func(*mml.Node) error
}

func notInAlignment(name string) error {
	return texError("NotInAlignment", "\\%s can only be used in aligment environments", name)
}

func (p *parser) arrowBetweenLines(name string) error {
	// GetStar uses GetNext's JavaScript whitespace, just like GetBrackets.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	star := p.pos < len(p.source) && p.source[p.pos] == '*'
	if star {
		p.pos++
	}
	defaultArrow := "\\Updownarrow"
	symbol, _, err := p.readBrackets(&defaultArrow)
	if err != nil {
		if e, ok := err.(*Error); ok && e.ID == "MissingCloseBracket" {
			return texError(e.ID, "Could not find closing ']' for argument to \\%s", name)
		}
		return err
	}
	cells := []*mml.Node{}
	expansion := symbol + "\\quad"
	if star {
		// These are direct EndEntry calls, so unlike an authored Entry
		// token they do not clear the array's lexical environment.
		cells = append(cells, node("mtd"), node("mtd"))
		expansion = "\\quad" + symbol
	}
	content, err := p.parseString(expansion)
	if err != nil {
		return err
	}
	// The row is emitted before the enclosing SetFont continuation returns,
	// so resolve its lexical font here rather than losing that environment.
	if p.activeFont != "" || p.fontExplicitEmpty {
		applyScopedMathVariant(content, p.activeFont)
	}
	cells = append(cells, node("mtd", content))
	return p.arrayCell.equation.endRow(node("mtr", cells...))
}

// parseEquationRow lets commands end the actual EqnArray row while the same
// cell parser keeps reading. This preserves SetFont and macro expansion state,
// and the callback finalizes the tag before any following input is parsed.
func (p *parser) parseEquationRow(cells []string, final bool, appendRow func(*mml.Node) error) error {
	var entries []*mml.Node
	state := &equationRowState{}
	state.endRow = func(row *mml.Node) error {
		if err := appendRow(row); err != nil {
			return err
		}
		entries = nil
		state.columns = 0
		return nil
	}
	for _, raw := range cells {
		cell := &arrayCellState{equation: state}
		content, err := p.parseStringWithStackArray(strings.TrimSpace(raw), p.ensureStackGlobal(), nil, cell)
		if err != nil {
			return err
		}
		entries = append(entries, node("mtd", content))
		state.columns = len(entries)
	}
	if final && omitFinalArrayRow(0, 1, entries) {
		return nil
	}
	return appendRow(node("mtr", entries...))
}

func isEquationArray(environment string) bool {
	switch environment {
	case "align", "align*", "aligned", "alignedat", "alignat", "alignat*",
		"split", "gather", "gather*", "gathered", "lgathered", "rgathered":
		return true
	}
	return false
}

func (p *parser) parseEquationTable(body, environment string) (*mml.Node, error) {
	state := p.amsTags()
	taggable := environment == "alignat" || environment == "alignat*"
	state.start(environment, taggable, environment == "alignat")
	defer state.end()
	rows := splitTable(body)
	var mrows []*mml.Node
	var tags []*mml.Node
	appendRow := func(row *mml.Node) error {
		mrows = append(mrows, row)
		tag, err := state.getTag(p)
		if err != nil {
			return err
		}
		tags = append(tags, tag)
		state.clearTag()
		return nil
	}
	for i, cells := range rows {
		if err := p.parseEquationRow(cells, i == len(rows)-1, appendRow); err != nil {
			return nil, err
		}
	}
	table := node("mtable", mrows...)
	if !strings.Contains(environment, "gather") {
		prefixRelationColumns(table)
	}
	for i, tag := range tags {
		if tag != nil {
			table.Children[i] = node("mlabeledtr", append([]*mml.Node{tag}, table.Children[i].Children...)...)
			table.Children[i].Parent = table
		}
	}
	return table, nil
}
