// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/base/{BaseItems,BaseMappings,BaseMethods}.ts,
// ams/{AmsItems,AmsMappings,AmsMethods}.ts,
// amscd/{AmsCdConfiguration,AmsCdMappings,AmsCdMethods}.ts,
// cases/CasesConfiguration.ts,
// empheq/{EmpheqConfiguration,EmpheqUtil}.ts,
// mathtools/{MathtoolsConfiguration,MathtoolsItems,MathtoolsMappings,
// MathtoolsMethods,MathtoolsUtil}.ts, and
// newcommand/{NewcommandItems,NewcommandMethods}.ts.

package tex

import (
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) beginEnvironment(name string) ([]*mml.Node, error) {
	environment, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	environment = strings.TrimSpace(environment)
	if environment == "" {
		return nil, texError("UnknownEnv", "Unknown environment '%s'", environment)
	}
	if nodes, handled, err := p.mathtoolsEnvironment(environment); handled {
		return nodes, err
	}
	if p.state.augmentedPackages {
		if nodes, handled, err := p.empheqEnvironment(environment); handled {
			return nodes, err
		}
	}
	if nodes, handled, err := p.physicsEnvironment(environment); handled {
		return nodes, err
	}
	if nodes, handled, err := p.casesEnvironment(environment); handled {
		return nodes, err
	}
	if nodes, handled, err := p.amscdEnvironment(environment); handled {
		return nodes, err
	}
	if nodes, handled, err := p.baseAMSEnvironment(environment); handled {
		return nodes, err
	}
	if nodes, handled, err := p.amsTagEnvironment(environment); handled {
		return nodes, err
	}

	if definition, ok := p.state.environments[environment]; ok {
		args := make([]string, 0, definition.arguments)
		if definition.optionalDefault != nil {
			arg, present, err := p.readBrackets(definition.optionalDefault)
			if err != nil {
				return nil, err
			}
			if !present {
				arg = *definition.optionalDefault
			}
			args = append(args, arg)
		}
		for len(args) < definition.arguments {
			arg, _, err := p.readArgument("begin{"+environment+"}", false)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		begin, err := substituteArguments(definition.begin, args)
		if err != nil {
			return nil, err
		}
		end, err := substituteArguments(definition.end, nil)
		if err != nil {
			return nil, err
		}
		body, err := p.captureEnvironment(environment)
		if err != nil {
			return nil, err
		}
		parsed, err := p.parseString(begin + body + end)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{parsed}, nil
	}

	columnSpec := ""
	if environment == "array" || environment == "subarray" || environment == "crampedsubarray" {
		columnSpec, _, err = p.readArgument("begin{"+environment+"}", false)
		if err != nil {
			return nil, err
		}
	}
	if strings.HasSuffix(environment, "*") && strings.Contains(environment, "matrix") {
		if alignment, _, bracketErr := p.readBrackets(nil); bracketErr != nil {
			return nil, bracketErr
		} else if alignment != "" {
			columnSpec = alignment
		}
	}
	if strings.Contains(environment, "alignat") {
		// AMS consumes the maximum pair count before the environment body; it
		// does not appear in the resulting MathML table.
		_, _, err = p.readArgument("begin{"+environment+"}", false)
		if err != nil {
			return nil, err
		}
	}
	body, err := p.captureEnvironment(environment)
	if err != nil {
		return nil, err
	}

	switch environment {
	case "equation", "equation*", "displaymath", "math", "split":
		contents, err := p.parseString(body)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{contents}, nil
	case "matrix", "matrix*", "smallmatrix", "smallmatrix*", "array", "subarray", "crampedsubarray",
		"pmatrix", "pmatrix*", "bmatrix", "bmatrix*", "Bmatrix", "Bmatrix*", "vmatrix", "vmatrix*", "Vmatrix", "Vmatrix*",
		"psmallmatrix", "psmallmatrix*", "bsmallmatrix", "bsmallmatrix*", "Bsmallmatrix", "Bsmallmatrix*", "vsmallmatrix", "vsmallmatrix*", "Vsmallmatrix", "Vsmallmatrix*":
		style := "T"
		if strings.Contains(environment, "small") || strings.Contains(environment, "subarray") {
			style = "S"
		}
		table, err := p.parseTable(body, style)
		if err != nil {
			return nil, err
		}
		if environment == "subarray" || environment == "crampedsubarray" {
			resetTableAttributes(table,
				"columnspacing", "0em",
				"rowspacing", "0.1em",
			)
			table.SetProperty("useHeight", false)
			table.SetProperty("scriptlevel", 1)
			// Array's S/S' style is explicit while inheritance runs; the
			// cleanAttributes postfilter then removes it because it equals the
			// mtable default, leaving no materialized displaystyle layer.
			table.Attributes.Set("displaystyle", false)
		}
		applyColumnSpec(table, columnSpec)
		open, close := matrixDelimiters(environment)
		if open != "" || close != "" {
			table = p.leftRightFenced(open, table, close, true)
		}
		return []*mml.Node{table}, nil
	case "cases", "dcases", "rcases", "drcases", "cases*", "dcases*", "rcases*", "drcases*":
		style := "T"
		if strings.HasPrefix(environment, "d") {
			style = "D"
		}
		table, err := p.parseTable(body, style)
		if err != nil {
			return nil, err
		}
		resetTableAttributes(table,
			"columnalign", "left left",
			"columnspacing", "1em",
			"rowspacing", ".2em",
		)
		// AMS Array's text style is explicit during inheritance.  The
		// clean-attributes postfilter removes it because it matches mtable's
		// default, leaving only scriptlevel in the inherited layer.
		table.Attributes.Set("displaystyle", false)
		open, close := "{", ""
		if strings.Contains(environment, "rcases") {
			open, close = "", "}"
		}
		return []*mml.Node{p.leftRightFenced(open, table, close, true)}, nil
	case "numcases", "subnumcases":
		left, _, err := p.readArgument("begin{"+environment+"}", true)
		if err != nil {
			return nil, err
		}
		table, err := p.parseTable(body, "T")
		if err != nil {
			return nil, err
		}
		if left != "" {
			prefix, err := p.parseString(left)
			if err != nil {
				return nil, err
			}
			return []*mml.Node{prefix, p.leftRightFenced("{", table, "", true)}, nil
		}
		return []*mml.Node{p.leftRightFenced("{", table, "", true)}, nil
	case "align", "align*", "alignat", "alignat*", "xalignat", "xalignat*", "xxalignat", "aligned", "alignedat", "gather", "gather*", "gathered", "multline", "multline*", "multlined", "lgathered", "rgathered", "spreadlines":
		table, err := p.parseTable(body, "D")
		if err != nil {
			return nil, err
		}
		if strings.Contains(environment, "gather") {
			resetTableAttributes(table,
				"displaystyle", true,
				"columnalign", "center",
				"columnspacing", "1em",
				"rowspacing", "3pt",
				"side", "right",
				"minlabelspacing", "0.8em",
			)
		} else if strings.Contains(environment, "multline") {
			finishAMSMultlineTable(table)
		} else {
			prefixRelationColumns(table)
			if isAMSXAlignAt(environment) {
				padded := environment != "xxalignat"
				padAMSXAlignAtRows(table, padded)
				finishAMSXAlignAtTable(table, padded)
			} else {
				resetTableAttributes(table,
					"displaystyle", true,
					"columnalign", "right left",
					"columnspacing", "0em",
					"rowspacing", "3pt",
				)
			}
		}
		return []*mml.Node{table}, nil
	case "CD":
		return p.parseCD(body)
	default:
		return nil, texError("UnknownEnv", "Unknown environment '%s'", environment)
	}
}

func isAMSXAlignAt(environment string) bool {
	return environment == "xalignat" || environment == "xalignat*" || environment == "xxalignat"
}

// padAMSXAlignAtRows ports FlalignItem.EndRow's padding pass. Xalignat uses
// fit columns before, between, and after equation pairs; xxalignat does not.
func padAMSXAlignAtRows(table *mml.Node, padded bool) {
	if !padded {
		return
	}
	for _, tableRow := range table.Children {
		cells := make([]*mml.Node, 0, len(tableRow.Children)+3)
		cells = append(cells, node("mtd", forcedRow(nil, true)))
		for i, cell := range tableRow.Children {
			cells = append(cells, cell)
			if i%2 == 1 {
				cells = append(cells, node("mtd", forcedRow(nil, true)))
			}
		}
		tableRow.SetChildren(cells)
	}
}

// finishAMSXAlignAtTable is AmsMethods.XalignAt followed by FlalignArray's
// exact arraydef insertion order. D2 uses right-side tags; zero-width labels
// set minlabelspacing to zero.
func finishAMSXAlignAtTable(table *mml.Node, padded bool) {
	columnAlignPattern, columnWidthPattern := []string{"right", "left", "center"}, []string{"auto", "auto", "fit"}
	if padded {
		columnAlignPattern, columnWidthPattern = []string{"center", "right", "left"}, []string{"fit", "auto", "auto"}
	}
	maxColumns := 0
	for _, row := range table.Children {
		if len(row.Children) > maxColumns {
			maxColumns = len(row.Children)
		}
	}
	resetTableAttributes(table,
		"width", "100%",
		"displaystyle", true,
		"columnalign", repeatAMSColumnPattern(columnAlignPattern, maxColumns),
		"columnspacing", "0em",
		"columnwidth", repeatAMSColumnPattern(columnWidthPattern, maxColumns),
		"rowspacing", "3pt",
		"side", "right",
		"minlabelspacing", "0",
		"data-width-includes-label", true,
	)
}

func repeatAMSColumnPattern(pattern []string, count int) string {
	if count == 0 {
		return strings.Join(pattern, " ")
	}
	columns := make([]string, count)
	for i := range columns {
		columns[i] = pattern[i%len(pattern)]
	}
	return strings.Join(columns, " ")
}

// finishAMSMultlineTable ports AmsMethods.Multline's arraydef and
// MultlineItem.EndTable's default first/last row alignment.
func finishAMSMultlineTable(table *mml.Node) {
	alignAMSMultlineCells(table)
	resetTableAttributes(table,
		"displaystyle", true,
		"rowspacing", ".5em",
		"columnspacing", "100%",
		"width", "100%",
		"side", "right",
		"minlabelspacing", "0.8em",
		"framespacing", "1em 0",
		"frame", "",
		"data-width-includes-label", true,
	)
}

func alignAMSMultlineCells(table *mml.Node) {
	if len(table.Children) != 0 {
		firstCell := table.Children[0].Children[0]
		if _, explicit := firstCell.Attributes.GetExplicit("columnalign"); !explicit {
			firstCell.Attributes.Set("columnalign", "left")
		}
		lastRow := table.Children[len(table.Children)-1]
		lastCell := lastRow.Children[len(lastRow.Children)-1]
		if _, explicit := lastCell.Attributes.GetExplicit("columnalign"); !explicit {
			lastCell.Attributes.Set("columnalign", "right")
		}
	}
}

func prefixRelationColumns(table *mml.Node) {
	for _, tableRow := range table.Children {
		for column, cell := range tableRow.Children {
			if column%2 == 0 || len(cell.Children) == 0 {
				continue
			}
			contents := cell.Children[0]
			if contents.Kind != "mrow" || !contents.Flags.Inferred || len(contents.Children) == 0 {
				continue
			}
			first := contents.Children[0]
			if first.Kind != "mo" || first.TeXClass != mml.TeXClassRel {
				continue
			}
			empty := node("mi")
			contents.SetChildren(append([]*mml.Node{empty}, contents.Children...))
		}
	}
}

func (p *parser) captureEnvironment(environment string) (string, error) {
	start := p.pos
	depth := 1
	for i := p.pos; i < len(p.source); {
		if p.source[i] != '\\' {
			_, size := utf8.DecodeRuneInString(p.source[i:])
			i += size
			continue
		}
		commandStart := i
		scanner := &parser{source: p.source, pos: i + 1}
		name := scanner.readControlSequence()
		if name != "begin" && name != "end" {
			i = scanner.pos
			continue
		}
		raw, _, err := scanner.readArgument(name, false)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(raw) != environment {
			i = scanner.pos
			continue
		}
		if name == "begin" {
			depth++
		} else {
			depth--
			if depth == 0 {
				body := p.source[start:commandStart]
				p.pos = scanner.pos
				return body, nil
			}
		}
		i = scanner.pos
	}
	return "", texError("MissingEnd", "Missing \\end{%s}", environment)
}

func (p *parser) parseTable(body, style string) (*mml.Node, error) {
	rows := splitTable(body)
	mtrNodes := make([]*mml.Node, 0, len(rows))
	for _, cells := range rows {
		if len(cells) == 1 && strings.TrimSpace(cells[0]) == "" && len(rows) > 1 {
			continue
		}
		mtdNodes := make([]*mml.Node, 0, len(cells))
		for _, cell := range cells {
			contents, err := p.parseString(strings.TrimSpace(cell))
			if err != nil {
				return nil, err
			}
			mtdNodes = append(mtdNodes, node("mtd", contents))
		}
		mtrNodes = append(mtrNodes, node("mtr", mtdNodes...))
	}
	table := node("mtable", mtrNodes...)
	table.Attributes.Set("columnspacing", "1em")
	table.Attributes.Set("rowspacing", "4pt")
	return table, nil
}

func resetTableAttributes(table *mml.Node, entries ...any) {
	for _, name := range table.Attributes.ExplicitNames() {
		table.Attributes.Explicit().Delete(name)
	}
	for i := 0; i < len(entries); i += 2 {
		table.Attributes.Set(entries[i].(string), entries[i+1])
	}
}

// repeatAMSEqnArrayDefinition ports EqnArrayItem.extendArray.  Multi-value
// column definitions repeat to the longest row and are then truncated to the
// exact number of columns (or inter-column gaps).  A single value remains a
// single value because MathML itself repeats it.
func repeatAMSEqnArrayDefinition(definition string, maximum int) string {
	values := strings.Fields(definition)
	if len(values) <= 1 {
		return definition
	}
	for len(values) < maximum {
		values = append(values, values...)
	}
	if maximum < len(values) {
		values = values[:maximum]
	}
	return strings.Join(values, " ")
}

// finishAMSEqnArrayTable applies the source map's initial EqnArray column
// definitions followed by EqnArrayItem.EndTable's repetition pass.
func finishAMSEqnArrayTable(table *mml.Node, environment string, maximum int) {
	spacing := "0em"
	switch environment {
	case "align", "align*", "aligned":
		spacing = "0em 2em"
	}
	resetTableAttributes(table,
		"displaystyle", true,
		"columnalign", repeatAMSEqnArrayDefinition("right left", maximum),
		"columnspacing", repeatAMSEqnArrayDefinition(spacing, maximum-1),
		"rowspacing", "3pt",
	)
}

func splitTable(body string) [][]string {
	rows := [][]string{{}}
	cellStart := 0
	depth := 0
	for i := 0; i < len(body); {
		switch body[i] {
		case '{':
			depth++
			i++
		case '}':
			if depth > 0 {
				depth--
			}
			i++
		case '\\':
			if i+1 < len(body) && body[i+1] == '\\' && depth == 0 {
				rows[len(rows)-1] = append(rows[len(rows)-1], body[cellStart:i])
				i += 2
				if i < len(body) && body[i] == '[' {
					if end := strings.IndexByte(body[i:], ']'); end >= 0 {
						i += end + 1
					}
				}
				rows = append(rows, []string{})
				cellStart = i
				continue
			}
			i++
			if i < len(body) {
				_, size := utf8.DecodeRuneInString(body[i:])
				i += size
			}
		case '&':
			if depth == 0 {
				rows[len(rows)-1] = append(rows[len(rows)-1], body[cellStart:i])
				cellStart = i + 1
			}
			i++
		default:
			_, size := utf8.DecodeRuneInString(body[i:])
			i += size
		}
	}
	rows[len(rows)-1] = append(rows[len(rows)-1], body[cellStart:])
	return rows
}

func splitTopLevel(value string, delimiter rune) []string {
	parts := []string{}
	start := 0
	depth := 0
	for i := 0; i < len(value); {
		r, size := utf8.DecodeRuneInString(value[i:])
		if r == '\\' {
			i += size
			if i < len(value) {
				_, nextSize := utf8.DecodeRuneInString(value[i:])
				i += nextSize
			}
			continue
		}
		if r == '{' {
			depth++
		} else if r == '}' && depth > 0 {
			depth--
		} else if r == delimiter && depth == 0 {
			parts = append(parts, value[start:i])
			start = i + size
		}
		i += size
	}
	return append(parts, value[start:])
}

func matrixDelimiters(environment string) (string, string) {
	environment = strings.TrimSuffix(environment, "*")
	switch environment {
	case "pmatrix", "psmallmatrix":
		return "(", ")"
	case "bmatrix", "bsmallmatrix":
		return "[", "]"
	case "Bmatrix", "Bsmallmatrix":
		return "{", "}"
	case "vmatrix", "vsmallmatrix":
		return "|", "|"
	case "Vmatrix", "Vsmallmatrix":
		return "‖", "‖"
	}
	return "", ""
}

func applyColumnSpec(table *mml.Node, specification string) {
	var aligns []string
	var lines []string
	for _, char := range specification {
		switch char {
		case 'l':
			aligns = append(aligns, "left")
		case 'c':
			aligns = append(aligns, "center")
		case 'r':
			aligns = append(aligns, "right")
		case '|':
			lines = append(lines, "solid")
		case ':':
			lines = append(lines, "dashed")
		}
	}
	if len(aligns) != 0 {
		table.Attributes.Set("columnalign", strings.Join(aligns, " "))
	}
	if len(lines) != 0 {
		table.Attributes.Set("columnlines", strings.Join(lines, " "))
	}
}

func (p *parser) displayLines(name string) ([]*mml.Node, error) {
	body, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	table, err := p.parseTable(strings.ReplaceAll(body, "&", "\\&"), "D")
	if err != nil {
		return nil, err
	}
	resetTableAttributes(table,
		"rowspacing", ".5em",
		"columnspacing", "1em",
		"displaystyle", true,
	)
	return []*mml.Node{table}, nil
}

func (p *parser) parseCD(body string) ([]*mml.Node, error) {
	// AmsCdMethods creates an mtable with alternating object and arrow cells.
	// Preserve the source's two-row cadence and expose @-arrow recipes as
	// stretchy relation operators; labels remain parsed TeX children.
	rows := splitCDRows(body)
	var mrows []*mml.Node
	for rowIndex, raw := range rows {
		cells, err := p.parseCDRow(raw, rowIndex)
		if err != nil {
			return nil, err
		}
		mrows = append(mrows, node("mtr", cells...))
	}
	table := node("mtable", mrows...)
	table.Attributes.Set("columnspacing", "5pt")
	table.Attributes.Set("rowspacing", "5pt")
	table.Attributes.Set("displaystyle", true)
	return []*mml.Node{table}, nil
}

func splitCDRows(body string) []string {
	rows := splitTable(body)
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, strings.Join(row, "&"))
	}
	return result
}

func (p *parser) parseCDRow(raw string, rowIndex int) ([]*mml.Node, error) {
	var cells []*mml.Node
	position := 0
	objectRow := rowIndex%2 == 0
	for position < len(raw) {
		at := strings.IndexByte(raw[position:], '@')
		if at < 0 {
			break
		}
		at += position
		prefix := strings.TrimSpace(raw[position:at])
		if prefix != "" {
			contents, err := p.parseString(prefix)
			if err != nil {
				return nil, err
			}
			if objectRow && len(cells) == 0 {
				contents = appendToRow(contents, cdStrut())
			}
			cells = appendCDCellContents(cells, contents, objectRow)
		}
		arrow, next, err := p.parseCDArrow(raw, at)
		if err != nil {
			return nil, err
		}
		cells = append(cells, node("mtd", arrow))
		if !objectRow {
			cells = append(cells, node("mtd"))
		}
		position = next
	}
	tail := strings.TrimSpace(raw[position:])
	if tail != "" {
		contents, err := p.parseString(tail)
		if err != nil {
			return nil, err
		}
		if objectRow && len(cells) == 0 {
			contents = appendToRow(contents, cdStrut())
		}
		cells = appendCDCellContents(cells, contents, objectRow)
	}
	if len(cells) == 0 {
		cells = append(cells, node("mtd"))
	}
	return cells, nil
}

// AmsCdMethods.arrow ends every arrow with a Cell item.  On an arrow row that
// leaves an active empty cell between alternating vertical-arrow columns.  TeX
// following the arrow fills that cell; only another immediate @ command (or
// the row end) commits it empty.  Preserve that stack behavior instead of
// appending the following object after the placeholder.
func appendCDCellContents(cells []*mml.Node, contents *mml.Node, objectRow bool) []*mml.Node {
	if !objectRow && len(cells) != 0 {
		last := cells[len(cells)-1]
		if last.Kind == "mtd" && len(last.Children) == 1 &&
			last.Children[0].Kind == "mrow" && last.Children[0].Flags.Inferred &&
			len(last.Children[0].Children) == 0 {
			cells[len(cells)-1] = node("mtd", contents)
			return cells
		}
	}
	return append(cells, node("mtd", contents))
}

func cdStrut() *mml.Node {
	strut := node("mpadded")
	strut.Attributes.Set("height", "8.5pt")
	strut.Attributes.Set("depth", "2pt")
	return strut
}

func appendToRow(contents, child *mml.Node) *mml.Node {
	if contents.Kind == "mrow" && contents.Flags.Inferred {
		contents.AppendChild(child)
		return contents
	}
	return forcedRow([]*mml.Node{contents, child}, true)
}

func (p *parser) parseCDArrow(raw string, at int) (*mml.Node, int, error) {
	if at+1 >= len(raw) {
		return nil, at, texError("CDMissingArrow", "Missing CD arrow after @")
	}
	kind := raw[at+1]
	if !strings.ContainsRune("><VA.|=", rune(kind)) {
		ordinary, err := p.parseString("@")
		return ordinary, at + 1, err
	}
	position := at + 2
	if kind == '.' {
		return forcedRow(nil, true), position, nil
	}
	if kind == '|' {
		return p.cdVerticalArrow("‖"), position, nil
	}
	if kind == '=' {
		return p.cdHorizontalArrow("="), position, nil
	}
	first, next, err := cdLabel(raw, position, kind)
	if err != nil {
		return nil, at, err
	}
	second, next, err := cdLabel(raw, next, kind)
	if err != nil {
		return nil, at, err
	}
	arrows := map[byte]string{'>': "→", '<': "←", 'V': "↓", 'A': "↑"}
	arrow := p.cdHorizontalArrow(arrows[kind])
	if kind == 'V' || kind == 'A' {
		arrow = p.cdVerticalArrow(arrows[kind])
	}
	if kind == '>' || kind == '<' {
		if first == "" {
			first = "\\kern " + p.amscdState(amscdMinWidthState, "2.75em")
		}
		over, err := p.parseString(first)
		if err != nil {
			return nil, at, err
		}
		over = cdArrowLabel(over, true)
		if second == "" {
			return node("mover", arrow, over), next, nil
		}
		under, err := p.parseString(second)
		if err != nil {
			return nil, at, err
		}
		under = cdArrowLabel(under, false)
		return node("munderover", arrow, under, over), next, nil
	}
	arrow.TeXClass = mml.TeXClassOrd
	if first == "" && second == "" {
		return arrow, next, nil
	}
	var children []*mml.Node
	if first != "" {
		label, err := p.parseString("\\scriptstyle\\llap{" + first + "}")
		if err != nil {
			return nil, at, err
		}
		children = append(children, label)
	}
	children = append(children, arrow)
	if second != "" {
		label, err := p.parseString("\\scriptstyle\\rlap{" + second + "}")
		if err != nil {
			return nil, at, err
		}
		children = append(children, label)
	}
	return forcedRow(children, false), next, nil
}

func cdArrowLabel(label *mml.Node, over bool) *mml.Node {
	padded := node("mpadded", label)
	padded.Attributes.Set("width", "+.67em")
	padded.Attributes.Set("lspace", ".33em")
	if over {
		padded.Attributes.Set("voffset", ".1em")
	}
	return padded
}

func cdHorizontalArrow(character string) *mml.Node {
	arrow := token("mo", character)
	arrow.Attributes.Set("minsize", "2.75em")
	arrow.Attributes.Set("stretchy", true)
	return arrow
}

func cdVerticalArrow(character string) *mml.Node {
	arrow := token("mo", character)
	arrow.Attributes.Set("minsize", "1.75em")
	arrow.Attributes.Set("stretchy", true)
	arrow.Attributes.Set("symmetric", true)
	arrow.Attributes.Set("lspace", 0)
	arrow.Attributes.Set("rspace", 0)
	return arrow
}

func cdLabel(raw string, start int, delimiter byte) (string, int, error) {
	depth := 0
	for index := start; index < len(raw); index++ {
		switch raw[index] {
		case '\\':
			index++
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		default:
			if raw[index] == delimiter && depth == 0 {
				return raw[start:index], index + 1, nil
			}
		}
	}
	return "", start, texError("CDMissingArrow", "Missing closing %c for CD arrow", delimiter)
}

func (p *parser) parseEmpheq(body string) ([]*mml.Node, error) {
	options, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	innerEnvRaw, _, err := p.readArgument("empheq", false)
	if err != nil {
		return nil, err
	}
	innerEnv := strings.TrimSpace(innerEnvRaw)
	inner := &parser{source: "\\begin{" + innerEnv + "}" + body + "\\end{" + innerEnv + "}", state: p.state, display: p.display}
	parsed, _, err := inner.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	content := row(parsed, true)
	settings := keyvalString(options)
	if left := settings["left"]; left != "" {
		open, _ := p.convertDelimiterArgument(left)
		content = p.fenced(open, content, "", true)
	}
	if right := settings["right"]; right != "" {
		close, _ := p.convertDelimiterArgument(right)
		content = p.fenced("", content, close, true)
	}
	return []*mml.Node{content}, nil
}
