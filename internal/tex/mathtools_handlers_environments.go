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

func (p *parser) mathtoolsAlignment(environment string) ([]*mml.Node, error) {
	verticalAlign, pairCount := "", ""
	if environment == "alignedat" {
		var err error
		verticalAlign, err = p.readAlignedatAlignment()
		if err != nil {
			return nil, err
		}
		pairCount, err = p.readEquationPairCount(environment)
		if err != nil {
			return nil, err
		}
	}
	if strings.Contains(environment, "alignat") {
		var err error
		pairCount, err = p.readEquationPairCount(environment)
		if err != nil {
			return nil, err
		}
	}
	if isGuardedEquationEnvironment(environment) {
		if err := p.checkEquationEnvironment(); err != nil {
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
	// Mathtools intercepts these arrays before the ordinary AMS handler.
	// Keep their special rows and the same row-local tag lifetime; alignedat
	// remains untaggable even when a Mathtools command triggers this route.
	var tagState *amsTagState
	var flalign *amsFlalignLayout
	if isAMSXAlignAt(environment) || environment == "flalign" || environment == "flalign*" {
		flalign = newAMSFlalignLayout(environment, pairCount)
		tagState = p.amsTags()
		numbered := environment == "xalignat" || environment == "flalign"
		tagState.start(environment, numbered, numbered)
		defer tagState.end()
	} else if environment == "gather" || environment == "gather*" || environment == "alignedat" {
		tagState = p.amsTags()
		tagState.start(environment, environment != "alignedat", environment == "gather")
		defer tagState.end()
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows)+1)
	rowTags := make(map[*mml.Node]*mml.Node)
	appendRows := func(rows ...*mml.Node) error {
		if flalign != nil {
			for _, row := range rows {
				if err := flalign.endEntry(len(row.Children)); err != nil {
					return err
				}
				prefixRelationColumns(node("mtable", row))
				flalign.endRow(row)
			}
		}
		if tagState != nil {
			tag, err := tagState.getTag(p)
			if err != nil {
				return err
			}
			if tag != nil {
				if flalign != nil {
					tag = flalign.label(tag)
				}
				rowTags[rows[0]] = tag
			}
			tagState.clearTag()
		}
		mrows = append(mrows, rows...)
		return nil
	}
	appendContinuation := func(raw string, force bool) error {
		row, err := p.mathtoolsPlainRow(splitTopLevel(strings.TrimSpace(raw), '&'))
		if err != nil {
			return err
		}
		// ArrowBetweenLines, ShortVDotsWithin and FlushSpaceBelow end a row
		// themselves. A following tag alone does not create an empty final row.
		if !force && !mathtoolsRowHasContent(row) {
			return nil
		}
		return appendRows(row)
	}
	adjustedRowSpacing := map[int]bool{}
	for rowIndex, cells := range rows {
		joined := strings.TrimSpace(strings.Join(cells, "&"))
		if strings.Contains(joined, "\\shortvdotswithin") {
			before, arg, rest, err := mathtoolsCommandParts(joined, "shortvdotswithin")
			if err != nil {
				return nil, err
			}
			if len(mrows) != 0 {
				adjustedRowSpacing[len(mrows)-1] = true
			}
			dotsRow := len(mrows)
			if tagState == nil {
				if err := appendRows(
					node("mtr", node("mtd"), node("mtd", p.mathtoolsVDots(arg, true))),
					node("mtr", node("mtd")),
				); err != nil {
					return nil, err
				}
			} else {
				prefix, err := p.mathtoolsContinuationNodes(before)
				if err != nil {
					return nil, err
				}
				if err := appendRows(node("mtr", node("mtd", prefix...), node("mtd", p.mathtoolsVDots(arg, true)))); err != nil {
					return nil, err
				}
				if err := appendContinuation(rest, rowIndex < len(rows)-1); err != nil {
					return nil, err
				}
			}
			adjustedRowSpacing[dotsRow] = true
			continue
		}
		if strings.Contains(joined, "\\vdotswithin") {
			before, arg, after, err := mathtoolsCommandParts(joined, "vdotswithin")
			if err != nil {
				return nil, err
			}
			flushAbove := strings.Contains(joined, "\\MTFlushSpaceAbove")
			if flushAbove && len(mrows) != 0 {
				adjustedRowSpacing[len(mrows)-1] = true
			}
			dotsRow := len(mrows)
			contents := []*mml.Node{}
			if tagState != nil {
				contents, err = p.mathtoolsContinuationNodes(strings.ReplaceAll(before, "\\MTFlushSpaceAbove", ""))
				if err != nil {
					return nil, err
				}
			}
			contents = append(contents, p.mathtoolsVDots(arg, flushAbove))
			flushSource := joined
			if tagState != nil {
				flushSource = after
				tail := after
				if flush := strings.Index(tail, "\\MTFlushSpaceBelow"); flush >= 0 {
					tail = tail[:flush]
				}
				suffix, err := p.mathtoolsContinuationNodes(tail)
				if err != nil {
					return nil, err
				}
				contents = append(contents, suffix...)
			}
			if err := appendRows(node("mtr", node("mtd", contents...))); err != nil {
				return nil, err
			}
			if flush := strings.Index(flushSource, "\\MTFlushSpaceBelow"); flush >= 0 {
				adjustedRowSpacing[dotsRow] = true
				rest := strings.TrimSpace(flushSource[flush+len("\\MTFlushSpaceBelow"):])
				if tagState != nil {
					if err := appendContinuation(rest, rowIndex < len(rows)-1); err != nil {
						return nil, err
					}
				} else if rest != "" {
					restCells := splitTopLevel(rest, '&')
					row, err := p.mathtoolsPlainRow(restCells)
					if err != nil {
						return nil, err
					}
					if err := appendRows(row); err != nil {
						return nil, err
					}
				}
			}
			continue
		}
		if len(cells) != 0 && strings.Contains(cells[len(cells)-1], "\\Aboxed") {
			row, err := p.mathtoolsAboxedRow(cells, tagState != nil)
			if err != nil {
				return nil, err
			}
			if err := appendRows(row); err != nil {
				return nil, err
			}
			continue
		}
		if isEquationArray(environment) {
			if err := p.parseEquationRow(cells, rowIndex == len(rows)-1, func(row *mml.Node) error { return appendRows(row) }); err != nil {
				return nil, err
			}
		} else {
			mtds := make([]*mml.Node, 0, len(cells))
			for _, raw := range cells {
				content, err := p.parseArrayCellString(strings.TrimSpace(raw))
				if err != nil {
					return nil, err
				}
				mtds = append(mtds, node("mtd", content))
			}
			if omitFinalArrayRow(rowIndex, len(rows), mtds) {
				continue
			}
			if err := appendRows(node("mtr", mtds...)); err != nil {
				return nil, err
			}
		}
	}
	table := node("mtable", mrows...)
	if tagState != nil {
		if flalign == nil {
			prefixEquationRelationColumns(table, 1)
		}
		for i, row := range table.Children {
			if tag := rowTags[row]; tag != nil {
				table.Children[i] = node("mlabeledtr", append([]*mml.Node{tag}, row.Children...)...)
				table.Children[i].Parent = table
			}
		}
	}
	if flalign != nil {
		flalign.endTable(table)
	} else if strings.Contains(environment, "gather") {
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
		maximumColumns := 0
		for _, row := range table.Children {
			if len(row.Children) > maximumColumns {
				maximumColumns = len(row.Children)
			}
		}
		finishAMSEqnArrayTable(table, environment, maximumColumns)
		if environment == "alignedat" {
			finishAlignedatTable(table, pairCount, verticalAlign, maximumColumns)
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

func (p *parser) mathtoolsVDots(argument string, flush bool) *mml.Node {
	arg, _ := p.parseString(argument)
	baseChildren := []*mml.Node{token("mi", "")}
	baseChildren = append(baseChildren, unwrapInferred(arg)...)
	baseChildren = append(baseChildren, token("mi", ""))
	dots := p.token("mo", "⋮")
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

func (p *parser) mathtoolsAboxedRow(cells []string, preserveContinuation bool) (*mml.Node, error) {
	mtds := make([]*mml.Node, 0, len(cells)+2)
	for _, raw := range cells[:len(cells)-1] {
		content, err := p.parseArrayCellString(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		mtds = append(mtds, node("mtd", content))
	}
	if len(mtds)%2 == 1 {
		mtds = append(mtds, node("mtd"))
	}
	raw := strings.TrimSpace(cells[len(cells)-1])
	before, argument, after, err := mathtoolsCommandParts(raw, "Aboxed")
	if err != nil {
		return nil, err
	}
	parts := splitTopLevel(argument, '&')
	left, right := parts[0], ""
	if len(parts) > 1 {
		right = parts[1]
	}
	if !preserveContinuation {
		before, after = "", ""
	}
	first, err := p.parseArrayCellString(before + "\\rlap{\\boxed{" + left + "{}" + right + "}}\\kern.267em\\phantom{" + left + "}")
	if err != nil {
		return nil, err
	}
	second, err := p.parseArrayCellString("\\phantom{{}" + right + "}\\kern.267em" + after)
	if err != nil {
		return nil, err
	}
	mtds = append(mtds, node("mtd", first), node("mtd", second))
	return node("mtr", mtds...), nil
}

func (p *parser) mathtoolsPlainRow(cells []string) (*mml.Node, error) {
	mtds := make([]*mml.Node, 0, len(cells))
	for _, raw := range cells {
		content, err := p.parseArrayCellString(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		mtds = append(mtds, node("mtd", content))
	}
	return node("mtr", mtds...), nil
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

func (p *parser) mathtoolsContinuationNodes(source string) ([]*mml.Node, error) {
	if strings.TrimSpace(source) == "" {
		return nil, nil
	}
	content, err := p.parseContinuationString(source)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(content), nil
}

func mathtoolsRowHasContent(row *mml.Node) bool {
	return !emptyArrayCells(row.Children)
}

func mathtoolsOnlyCommandArgument(source, command string) (string, error) {
	return mathtoolsCommandArgument(strings.TrimSpace(source), command)
}
