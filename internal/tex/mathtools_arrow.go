// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts, MathtoolsUtil.ts,
// and base/BaseItems.ts (EqnArrayItem and ArrayItem.addRowSpacing).
package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

// Only EqnArrayItem owns this state. FlalignItem has a different kind even
// though its implementation extends EqnArrayItem.
type equationRowState struct {
	entries []*mml.Node
	table   *equationTableState
}

type equationTableState struct {
	rows       int
	flushAbove int
	spacing    []string
	appendRow  func(*mml.Node) error
}

func newEquationTableState(appendRow func(*mml.Node) error) *equationTableState {
	return &equationTableState{flushAbove: -1, appendRow: appendRow}
}

func (state *equationRowState) endEntry(nodes []*mml.Node) {
	state.entries = append(state.entries, node("mtd", matrixCellContent(nodes)))
}

func (state *equationRowState) endRow() error {
	if err := state.table.appendRow(node("mtr", state.entries...)); err != nil {
		return err
	}
	state.entries = nil
	state.table.rows++
	return nil
}

func (state *equationTableState) addSpacing(adjust string) {
	if state.spacing == nil {
		state.spacing = []string{"3pt"}
	}
	// The source caches the original spacing, not a previous adjustment.
	for len(state.spacing) < state.rows {
		state.spacing = append(state.spacing, "0.3em")
	}
	if state.rows != 0 {
		spacing := 0.3 + matrixDimensionEm(adjust)
		if spacing < 0 {
			spacing = 0
		}
		state.spacing[state.rows-1] = layout.Em(spacing)
	}
}

func (state *equationTableState) applySpacing(table *mml.Node) {
	if state != nil && state.spacing != nil {
		// ArrayItem.checkLines fills trailing rows with the cached original
		// spacing, so a shortened last explicit value is not repeated.
		for len(state.spacing) < state.rows {
			state.spacing = append(state.spacing, "0.3em")
		}
		table.Attributes.Set("rowspacing", strings.Join(state.spacing, " "))
	}
}

func notInAlignment(name string) error {
	return texError("NotInAlignment", "\\%s can only be used in aligment environments", name)
}

func isMathtoolsArrayCommand(name string) bool {
	switch name {
	case "ArrowBetweenLines", "Aboxed", "vdotswithin", "shortvdotswithin", "MTFlushSpaceAbove", "MTFlushSpaceBelow":
		return true
	}
	return false
}

func (p *parser) readMathtoolsStar() bool {
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	star := p.pos < len(p.source) && p.source[p.pos] == '*'
	if star {
		p.pos++
	}
	return star
}

func (p *parser) readMathtoolsArgument(name string) (string, error) {
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	argument, _, err := p.readArgumentAtCursor(name, false)
	return argument, err
}

func (p *parser) arrowBetweenLines(name string) error {
	star := p.readMathtoolsStar()
	defaultArrow := "\\Updownarrow"
	symbol, _, err := p.readBrackets(&defaultArrow)
	if err != nil {
		if e, ok := err.(*Error); ok && e.ID == "MissingCloseBracket" {
			return texError(e.ID, "Could not find closing ']' for argument to \\%s", name)
		}
		return err
	}
	state := p.arrayCell.equation
	expansion := symbol + "\\quad"
	if star {
		// Direct EndEntry does not clear the lexical environment.
		state.endEntry(nil)
		state.endEntry(nil)
		expansion = "\\quad" + symbol
	}
	content, err := p.parseChild(expansion)
	if err != nil {
		return err
	}
	if p.activeFont != "" || p.fontExplicitEmpty {
		applyScopedMathVariant(content, p.activeFont)
	}
	state.endEntry(unwrapInferred(content))
	return state.endRow()
}

func (p *parser) equationAboxed(name string) error {
	state := p.arrayCell.equation
	if len(state.entries)%2 == 1 {
		state.entries = append(state.entries, node("mtd"))
	}
	argument, err := p.readMathtoolsArgument(name)
	if err != nil {
		return err
	}
	parts := splitTopLevel(argument, '&')
	left, right := parts[0], ""
	if len(parts) > 1 {
		right = parts[1]
	}
	expansion := "\\rlap{\\boxed{" + left + "{}" + right + "}}\\kern.267em\\phantom{" + left + "}&\\phantom{{}" + right + "}\\kern.267em"
	p.source, p.pos = expansion+p.source[p.pos:], 0
	return nil
}

func (p *parser) equationVDots(name string, flush bool) (*mml.Node, error) {
	argument, err := p.readMathtoolsArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseChild("\\mmlToken{mi}{}" + argument + "\\mmlToken{mi}{}")
	if err != nil {
		return nil, err
	}
	inner := node("mpadded", p.token("mo", "⋮"))
	inner.Attributes.Set("width", 0)
	inner.Attributes.Set("lspace", "-.5width")
	if flush {
		inner.Attributes.Set("height", "-.6em")
		inner.Attributes.Set("voffset", "-.18em")
	}
	outer := node("mpadded", inner, node("mphantom", base))
	outer.Attributes.Set("lspace", ".5width")
	return outer, nil
}

// Authored and macro-generated Entry tokens use the same parser boundary as
// CD. Readers can consume an ampersand as an argument without delivering one.
// Direct EndEntry/EndRow commands keep reading in that parser, preserving its
// font environment; an actual Entry starts a cell with the cleared Array env.
func (p *parser) parseEquationRow(cells []string, final bool, appendRow func(*mml.Node) error, tableState ...*equationTableState) error {
	table := newEquationTableState(appendRow)
	if len(tableState) != 0 {
		table = tableState[0]
	}
	state := &equationRowState{table: table}
	source := strings.Join(cells, "&")
	for {
		sub := p.matrixCellParser(source)
		sub.matrixClose = false
		sub.cdArrayEntry = true
		sub.arrayCell.equation = state
		children, _, err := sub.parseRowWithInfix(0, false, false)
		if err != nil {
			return err
		}
		if !sub.cdEntryStopped && final && len(children) == 0 && len(state.entries) == 0 {
			return nil
		}
		state.endEntry(children)
		if !sub.cdEntryStopped {
			return state.endRow()
		}
		source = sub.source[sub.pos:]
	}
}

func isEquationArray(environment string) bool {
	switch environment {
	case "align", "align*", "aligned", "alignedat", "alignat", "alignat*", "split", "gather", "gather*", "gathered", "lgathered", "rgathered":
		return true
	}
	return false
}

func (p *parser) parseEquationTable(body, environment string) (*mml.Node, *equationTableState, error) {
	state := p.amsTags()
	taggable := environment == "alignat" || environment == "alignat*"
	state.start(environment, taggable, environment == "alignat")
	defer state.end()
	rows := splitTable(body)
	var mrows, tags []*mml.Node
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
	spacing := newEquationTableState(appendRow)
	for i, cells := range rows {
		if err := p.parseEquationRow(cells, i == len(rows)-1, appendRow, spacing); err != nil {
			return nil, nil, err
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
	return table, spacing, nil
}
