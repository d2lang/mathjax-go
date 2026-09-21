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
		"\\Aboxed", "\\ArrowBetweenLines", "\\vdotswithin", "\\shortvdotswithin",
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
		table = leftRightFenced(open, table, close, true)
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
		content, err := p.parseString(raw)
		if err != nil {
			return nil, err
		}
		children := unwrapInferred(content)
		cell := node("mtd", children...)
		if align != "center" {
			cell.Attributes.Set("columnalign", align)
		}
		mrows = append(mrows, node("mtr", cell))
		_ = i
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
	spread, err := p.readDimension("begin{" + environment + "}")
	if err != nil {
		return nil, err
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	content, err := p.parseString(body)
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
	for _, cells := range rows {
		mtds := make([]*mml.Node, 0, len(cells))
		for i, raw := range cells {
			if i == 1 {
				text := node("mstyle", node("mtext", mml.NewText(strings.TrimSpace(raw))))
				mtds = append(mtds, node("mtd", text))
				continue
			}
			content, err := p.parseString(strings.TrimSpace(raw))
			if err != nil {
				return nil, err
			}
			mtds = append(mtds, node("mtd", content))
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
	return []*mml.Node{leftRightFenced(open, table, close, true)}, nil
}

func (p *parser) mathtoolsAlignment(environment string) ([]*mml.Node, error) {
	if strings.Contains(environment, "alignat") {
		if _, _, err := p.readArgument("begin{"+environment+"}", false); err != nil {
			return nil, err
		}
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}
	if strings.Contains(environment, "multline") {
		return p.mathtoolsMultlineBody(body)
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows)+1)
	adjustedRowSpacing := map[int]bool{}
	for _, cells := range rows {
		joined := strings.TrimSpace(strings.Join(cells, "&"))
		if strings.HasPrefix(joined, "\\ArrowBetweenLines") {
			arrowRows, err := p.mathtoolsArrowBetweenLines(joined)
			if err != nil {
				return nil, err
			}
			mrows = append(mrows, arrowRows...)
			continue
		}
		if strings.Contains(joined, "\\shortvdotswithin") {
			arg, err := mathtoolsCommandArgument(joined, "shortvdotswithin")
			if err != nil {
				return nil, err
			}
			if len(mrows) != 0 {
				adjustedRowSpacing[len(mrows)-1] = true
			}
			dotsRow := len(mrows)
			mrows = append(mrows,
				node("mtr", node("mtd"), node("mtd", p.mathtoolsVDots(arg, true))),
				node("mtr", node("mtd")),
			)
			adjustedRowSpacing[dotsRow] = true
			continue
		}
		if strings.Contains(joined, "\\vdotswithin") {
			arg, err := mathtoolsCommandArgument(joined, "vdotswithin")
			if err != nil {
				return nil, err
			}
			flushAbove := strings.Contains(joined, "\\MTFlushSpaceAbove")
			if flushAbove && len(mrows) != 0 {
				adjustedRowSpacing[len(mrows)-1] = true
			}
			dotsRow := len(mrows)
			mrows = append(mrows, node("mtr", node("mtd", p.mathtoolsVDots(arg, flushAbove))))
			if flush := strings.Index(joined, "\\MTFlushSpaceBelow"); flush >= 0 {
				adjustedRowSpacing[dotsRow] = true
				rest := strings.TrimSpace(joined[flush+len("\\MTFlushSpaceBelow"):])
				if rest != "" {
					restCells := splitTopLevel(rest, '&')
					row, err := p.mathtoolsPlainRow(restCells)
					if err != nil {
						return nil, err
					}
					mrows = append(mrows, row)
				}
			}
			continue
		}
		if len(cells) != 0 && strings.Contains(cells[len(cells)-1], "\\Aboxed") {
			row, err := p.mathtoolsAboxedRow(cells)
			if err != nil {
				return nil, err
			}
			mrows = append(mrows, row)
			continue
		}
		mtds := make([]*mml.Node, 0, len(cells))
		for _, raw := range cells {
			content, err := p.parseString(strings.TrimSpace(raw))
			if err != nil {
				return nil, err
			}
			mtds = append(mtds, node("mtd", content))
		}
		mrows = append(mrows, node("mtr", mtds...))
	}
	table := node("mtable", mrows...)
	if strings.Contains(environment, "gather") {
		resetTableAttributes(table,
			"displaystyle", true,
			"columnalign", "center",
			"columnspacing", "1em",
			"rowspacing", "3pt",
			"side", "right",
			"minlabelspacing", "0.8em",
		)
	} else {
		prefixRelationColumns(table)
		if isAMSXAlignAt(environment) {
			padded := environment != "xxalignat"
			padAMSXAlignAtRows(table, padded)
			finishAMSXAlignAtTable(table, padded)
		} else {
			maximumColumns := 0
			for _, row := range table.Children {
				if len(row.Children) > maximumColumns {
					maximumColumns = len(row.Children)
				}
			}
			finishAMSEqnArrayTable(table, environment, maximumColumns)
		}
	}
	if len(adjustedRowSpacing) != 0 {
		rowSpacing := make([]string, len(mrows))
		for i := range rowSpacing {
			rowSpacing[i] = "0.3em"
		}
		if len(rowSpacing) != 0 {
			rowSpacing[0] = "3pt"
		}
		for row := range adjustedRowSpacing {
			rowSpacing[row] = "0.1em"
		}
		table.Attributes.Set("rowspacing", strings.Join(rowSpacing, " "))
	}
	return []*mml.Node{table}, nil
}

func (p *parser) mathtoolsMultlineBody(body string) ([]*mml.Node, error) {
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for _, cells := range rows {
		raw := strings.TrimSpace(strings.Join(cells, "&"))
		shove := ""
		if strings.HasPrefix(raw, "\\shoveleft") {
			shove = "left"
			raw, _ = mathtoolsOnlyCommandArgument(raw, "shoveleft")
		} else if strings.HasPrefix(raw, "\\shoveright") {
			shove = "right"
			raw, _ = mathtoolsOnlyCommandArgument(raw, "shoveright")
		}
		content, err := p.parseString(raw)
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
		mrows = append(mrows, node("mtr", cell))
	}
	table := node("mtable", mrows...)
	finishAMSMultlineTable(table)
	return []*mml.Node{table}, nil
}

func (p *parser) mathtoolsArrowBetweenLines(raw string) ([]*mml.Node, error) {
	scanner := &parser{source: raw, state: p.state, display: p.display}
	scanner.skipSpaces()
	scanner.pos++
	name := scanner.readControlSequence()
	star := scanner.readStar()
	defaultArrow := "\\Updownarrow"
	symbol, _, err := scanner.readBrackets(&defaultArrow)
	if err != nil {
		return nil, err
	}
	expansion := symbol + "\\quad"
	if star {
		expansion = "\\quad" + symbol
	}
	content, err := p.parseString(symbol)
	if err != nil {
		return nil, err
	}
	space := node("mstyle", mathtoolsSpace("1em"))
	children := unwrapInferred(content)
	if star {
		children = append([]*mml.Node{space}, children...)
	} else {
		children = append(children, space)
	}
	content = forcedRow(children, true)
	_ = name
	_ = expansion
	return []*mml.Node{node("mtr", node("mtd", content)), node("mtr", node("mtd"))}, nil
}

func (p *parser) mathtoolsVDots(argument string, flush bool) *mml.Node {
	arg, _ := p.parseString(argument)
	baseChildren := []*mml.Node{token("mi", "")}
	baseChildren = append(baseChildren, unwrapInferred(arg)...)
	baseChildren = append(baseChildren, token("mi", ""))
	dots := token("mo", "⋮")
	inner := node("mpadded", dots)
	inner.Attributes.Set("width", 0)
	inner.Attributes.Set("lspace", "-.5width")
	if flush {
		inner.Attributes.Set("height", "-.6em")
		inner.Attributes.Set("voffset", "-.18em")
	}
	outer := node("mpadded", inner, node("mphantom", forcedRow(baseChildren, true)))
	outer.Attributes.Set("lspace", ".5width")
	return outer
}

func (p *parser) mathtoolsAboxedRow(cells []string) (*mml.Node, error) {
	mtds := make([]*mml.Node, 0, len(cells)+2)
	for _, raw := range cells[:len(cells)-1] {
		content, err := p.parseString(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		mtds = append(mtds, node("mtd", content))
	}
	if len(mtds)%2 == 1 {
		mtds = append(mtds, node("mtd"))
	}
	raw := strings.TrimSpace(cells[len(cells)-1])
	argument, err := mathtoolsCommandArgument(raw, "Aboxed")
	if err != nil {
		return nil, err
	}
	parts := splitTopLevel(argument, '&')
	left, right := parts[0], ""
	if len(parts) > 1 {
		right = parts[1]
	}
	first, err := p.parseString("\\rlap{\\boxed{" + left + "{}" + right + "}}\\kern.267em\\phantom{" + left + "}")
	if err != nil {
		return nil, err
	}
	second, err := p.parseString("\\phantom{{}" + right + "}\\kern.267em")
	if err != nil {
		return nil, err
	}
	mtds = append(mtds, node("mtd", first), node("mtd", second))
	return node("mtr", mtds...), nil
}

func (p *parser) mathtoolsPlainRow(cells []string) (*mml.Node, error) {
	mtds := make([]*mml.Node, 0, len(cells))
	for _, raw := range cells {
		content, err := p.parseString(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		mtds = append(mtds, node("mtd", content))
	}
	return node("mtr", mtds...), nil
}

func mathtoolsCommandArgument(source, command string) (string, error) {
	index := strings.Index(source, "\\"+command)
	if index < 0 {
		return "", texError("MissingArgFor", "Missing argument for \\%s", command)
	}
	scanner := &parser{source: source, pos: index + len(command) + 1}
	argument, _, err := scanner.readArgument(command, false)
	return argument, err
}

func mathtoolsOnlyCommandArgument(source, command string) (string, error) {
	return mathtoolsCommandArgument(strings.TrimSpace(source), command)
}
