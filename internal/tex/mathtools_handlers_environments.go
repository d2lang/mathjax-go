// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts,
// MathtoolsItems.ts, and MathtoolsUtil.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func mathtoolsEnvironmentHasSpecial(source string) bool {
	for _, command := range []string{
		"\\Aboxed", "\\vdotswithin", "\\shortvdotswithin",
		"\\MTFlushSpaceAbove", "\\MTFlushSpaceBelow", "\\shoveleft", "\\shoveright",
	} {
		if strings.Contains(source, command) {
			return true
		}
	}
	return false
}

// mathtoolsSmallMatrix ports MtSmallMatrix's Array arguments.  The starred
// forms read their alignment option; the fixed non-starred delimiter forms
// use the source-map's literal center alignment.
func (p *parser) mathtoolsSmallMatrix(environment string) ([]*mml.Node, error) {
	alignment := "c"
	if strings.HasSuffix(environment, "*") {
		defaultAlignment := p.mathtoolsOption("smallmatrix-align")
		var err error
		alignment, _, err = p.readBrackets(&defaultAlignment)
		if err != nil {
			return nil, err
		}
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	table, err := p.parseTable(body, "S")
	if err != nil {
		return nil, err
	}
	resetTableAttributes(table,
		"columnalign", "center",
		"columnspacing", ".333em",
		"rowspacing", ".2em",
		"displaystyle", false,
	)
	applyColumnSpec(table, alignment)
	table.SetProperty("useHeight", false)
	table.SetProperty("scriptlevel", 1)
	open, close := matrixDelimiters(environment)
	if open != "" || close != "" {
		table = p.leftRightFenced(open, table, close, true)
	}
	return []*mml.Node{table}, nil
}

func (p *parser) mathtoolsMultlined(environment string) ([]*mml.Node, error) {
	defaultPosition := p.mathtoolsOption("multlined-pos")
	position, present, err := p.readBrackets(&defaultPosition)
	if err != nil {
		return nil, err
	}
	width := ""
	if position != "" {
		width, _, err = p.readBrackets(nil)
		if err != nil {
			return nil, err
		}
	}
	if present && !strings.Contains("cbt", position) {
		width, position = position, width
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for i, cells := range rows {
		raw := strings.TrimSpace(strings.Join(cells, "&"))
		align := "center"
		if strings.HasPrefix(raw, "\\shoveleft") {
			align = "left"
			raw, err = mathtoolsOnlyCommandArgument(raw, "shoveleft")
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(raw, "\\shoveright") {
			align = "right"
			raw, err = mathtoolsOnlyCommandArgument(raw, "shoveright")
			if err != nil {
				return nil, err
			}
		}
		content, err := p.parseArrayCellString(raw)
		if err != nil {
			return nil, err
		}
		children := unwrapInferred(content)
		if i != 0 {
			children = fixInitialMO(children)
		}
		cell := node("mtd", children...)
		if align != "center" {
			cell.Attributes.Set("columnalign", align)
		}
		if omitFinalArrayRow(i, len(rows), []*mml.Node{cell}) {
			continue
		}
		mrows = append(mrows, node("mtr", cell))
	}
	if len(mrows) > 1 {
		firstCell := mrows[0].Children[0]
		if value, _ := firstCell.Attributes.Get("columnalign"); value != "right" {
			firstSkip := p.mathtoolsOption("firstline-afterskip")
			if firstSkip == "" {
				firstSkip = p.mathtoolsOption("multlinegap")
			}
			mathtoolsAppendCell(firstCell, mathtoolsSpace(firstSkip), false)
		}
		lastCell := mrows[len(mrows)-1].Children[0]
		if value, _ := lastCell.Attributes.Get("columnalign"); value != "left" {
			lastSkip := p.mathtoolsOption("lastline-preskip")
			if lastSkip == "" {
				lastSkip = p.mathtoolsOption("multlinegap")
			}
			mathtoolsAppendCell(lastCell, mathtoolsSpace(lastSkip), true)
		}
	}
	table := node("mtable", mrows...)
	alignAMSMultlineCells(table)
	table.Attributes.Set("displaystyle", true)
	table.Attributes.Set("rowspacing", ".5em")
	if width == "" {
		width = "auto"
	}
	table.Attributes.Set("width", width)
	table.Attributes.Set("columnwidth", "100%")
	if position == "t" {
		table.Attributes.Set("align", "baseline 1")
	} else if position == "b" {
		table.Attributes.Set("align", "baseline -1")
	} else {
		table.Attributes.Set("align", "axis")
	}
	return []*mml.Node{table}, nil
}

func mathtoolsAppendCell(cell, child *mml.Node, prepend bool) {
	if len(cell.Children) == 0 {
		cell.SetChildren([]*mml.Node{forcedRow([]*mml.Node{child}, true)})
		return
	}
	content := cell.Children[0]
	children := unwrapInferred(content)
	if prepend {
		children = append([]*mml.Node{child}, children...)
	} else {
		children = append(children, child)
	}
	cell.SetChildren([]*mml.Node{forcedRow(children, true)})
}

func (p *parser) mathtoolsSpreadLines(environment string) ([]*mml.Node, error) {
	// GetDimen reports parser.currentCS (\begin), not its descriptive name.
	spread, err := p.readDimension("begin")
	if err != nil {
		return nil, err
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	content, err := p.parseContinuationString(body)
	if err != nil {
		return nil, err
	}
	content.Walk(func(current *mml.Node) bool {
		if current.Kind == "mtable" {
			mathtoolsAddRowSpacing(current, spread)
		}
		return true
	})
	return unwrapInferred(content), nil
}

// mathtoolsAddRowSpacing ports MathtoolsUtil.addRowSpacing.  Spreadlines adds
// its dimension to each existing table-row spacing rather than replacing the
// package-specific base spacing.
func mathtoolsAddRowSpacing(table *mml.Node, spread string) {
	value, _ := table.Attributes.Get("rowspacing")
	rows := strings.Fields(sourceValueString(value))
	if len(rows) == 0 {
		rows = []string{"1ex"}
	}
	delta := layout.Length2Em(spread, 0, 1, 16)
	for i, row := range rows {
		spacing := layout.Length2Em(row, 0, 1, 16) + delta
		if spacing < 0 {
			spacing = 0
		}
		rows[i] = layout.Em(spacing)
	}
	table.Attributes.Set("rowspacing", strings.Join(rows, " "))
}

func (p *parser) mathtoolsCases(environment string) ([]*mml.Node, error) {
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for rowIndex, cells := range rows {
		mtds := make([]*mml.Node, 0, len(cells))
		for i, raw := range cells {
			if i == 1 {
				text := node("mstyle", node("mtext", mml.NewText(strings.TrimSpace(raw))))
				mtds = append(mtds, node("mtd", text))
				continue
			}
			content, err := p.parseArrayCellString(strings.TrimSpace(raw))
			if err != nil {
				return nil, err
			}
			mtds = append(mtds, arrayCellNode(content))
		}
		if omitFinalArrayRow(rowIndex, len(rows), mtds) {
			continue
		}
		mrows = append(mrows, node("mtr", mtds...))
	}
	table := node("mtable", mrows...)
	table.Attributes.Set("rowspacing", ".2em")
	table.Attributes.Set("columnspacing", "1em")
	table.Attributes.Set("columnalign", "left")
	if strings.HasPrefix(environment, "d") {
		table.Attributes.Set("displaystyle", true)
	}
	open, close := "{", ""
	if strings.Contains(environment, "rcases") {
		open, close = "", "}"
	}
	return []*mml.Node{p.leftRightFenced(open, table, close, true)}, nil
}

func (p *parser) mathtoolsMultline(environment string) ([]*mml.Node, error) {
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, err
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	return p.mathtoolsMultlineBody(body)
}

func (p *parser) mathtoolsMultlineBody(body string) ([]*mml.Node, error) {
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for rowIndex, cells := range rows {
		raw := strings.TrimSpace(strings.Join(cells, "&"))
		shove := ""
		if strings.HasPrefix(raw, "\\shoveleft") {
			shove = "left"
			raw, _ = mathtoolsOnlyCommandArgument(raw, "shoveleft")
		} else if strings.HasPrefix(raw, "\\shoveright") {
			shove = "right"
			raw, _ = mathtoolsOnlyCommandArgument(raw, "shoveright")
		}
		content, err := p.parseArrayCellString(raw)
		if err != nil {
			return nil, err
		}
		if shove != "" {
			content = texAtom(content, mml.TeXClassOrd)
		}
		cell := node("mtd", content)
		if shove != "" {
			cell.Attributes.Set("columnalign", shove)
		}
		if omitFinalArrayRow(rowIndex, len(rows), []*mml.Node{cell}) {
			continue
		}
		mrows = append(mrows, node("mtr", cell))
	}
	table := node("mtable", mrows...)
	finishAMSMultlineTable(table)
	return []*mml.Node{table}, nil
}

func mathtoolsCommandArgument(source, command string) (string, error) {
	_, argument, _, err := mathtoolsCommandParts(source, command)
	return argument, err
}

func mathtoolsCommandParts(source, command string) (string, string, string, error) {
	index := strings.Index(source, "\\"+command)
	if index < 0 {
		return "", "", "", texError("MissingArgFor", "Missing argument for \\%s", command)
	}
	scanner := &parser{source: source, pos: index + len(command) + 1}
	argument, _, err := scanner.readArgument(command, false)
	if err != nil {
		return "", "", "", err
	}
	return source[:index], argument, source[scanner.pos:], nil
}

func mathtoolsOnlyCommandArgument(source, command string) (string, error) {
	return mathtoolsCommandArgument(strings.TrimSpace(source), command)
}
