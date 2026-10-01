// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/base/{BaseMethods,BaseItems}.ts.

package tex

import (
	"strconv"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) matrixCommand(name string) ([]*mml.Node, error) {
	numbered := name == "eqalignno" || name == "leqalignno"
	aligned := name == "eqalign" || numbered
	displayLines := name == "displaylines"
	var body string
	var err error
	if displayLines {
		// Matrix installs a required-close Array at the current cursor.
		// Run macros before deciding which close reaches that Array.
		err = p.startMatrixBody(name)
	} else {
		body, err = p.readMatrixBody(name)
	}
	if err != nil {
		return nil, err
	}
	table := node("mtable")
	var liveArray *ordinaryArrayItem
	if displayLines {
		// Standalone Matrix has no Begin to continue after a direct Pop.
		liveArray = &ordinaryArrayItem{}
	}
	var entries []*mml.Node
	spacing := &arrayRowSpacing{}
	owner := arrayBodyOwner{
		live:         displayLines,
		ordinary:     liveArray,
		rules:        newArrayRules(table),
		requireClose: true,
		hasEntries:   func() bool { return len(entries) != 0 },
		endEntry: func(children []*mml.Node, fill *arrayCellState) error {
			content := fill.finish(matrixCellContent(children), len(children))
			entries = append(entries, arrayCellNode(content))
			return nil
		},
		endRow: func() error {
			line := node("mtr", entries...)
			if numbered && len(entries) == 3 {
				line = node("mlabeledtr", entries[2], entries[0], entries[1])
			}
			table.AppendChild(line)
			entries = nil
			spacing.rows++
			return nil
		},
		addSpacing: spacing.add,
	}
	if name == "cases" {
		owner.prepare = func(sub *parser) ([]*mml.Node, error) {
			if len(entries) != 1 {
				return nil, nil
			}
			return sub.prepareMatrixCasesText()
		}
	}
	if err := p.parseArrayBody(body, owner); err != nil {
		return nil, err
	}
	if liveArray != nil && liveArray.popped {
		return liveArray.nodes, nil
	}
	table.Attributes.Set("rowspacing", "4pt")
	table.Attributes.Set("columnspacing", "1em")
	if aligned {
		table.Attributes.Set("rowspacing", ".5em")
		table.Attributes.Set("columnspacing", "0.278em")
		table.Attributes.Set("displaystyle", true)
		table.Attributes.Set("columnalign", "right left")
	}
	if displayLines {
		table.Attributes.Set("rowspacing", ".5em")
		table.Attributes.Set("displaystyle", true)
		table.Attributes.Set("columnalign", "center")
	}
	if numbered {
		side := "right"
		if name == "leqalignno" {
			side = "left"
		}
		table.Attributes.Set("side", side)
	}
	if name == "cases" {
		table.Attributes.Set("rowspacing", ".1em")
		table.Attributes.Set("columnalign", "left left")
	}
	initial := "4pt"
	if aligned || displayLines {
		initial = ".5em"
	}
	if name == "cases" {
		initial = ".1em"
	}
	spacing.apply(table, initial)
	table = finishArrayRules(table)
	if name == "pmatrix" {
		return []*mml.Node{p.fenced("(", table, ")", true)}, nil
	}
	if name == "cases" {
		return []*mml.Node{p.fenced("{", table, "", true)}, nil
	}
	return []*mml.Node{table}, nil
}

// ArrayItem.addRowSpacing uses ParseUtil.dimen2em, whose TeX conversions
// differ from the SVG renderer's CSS lengths (notably ex and physical units).
func matrixDimensionEm(dimension string) float64 {
	match := dimensionFull.FindStringSubmatch(dimension)
	if match == nil {
		return 0
	}
	value, _ := strconv.ParseFloat(match[1], 64)
	switch match[2] {
	case "em":
		return value
	case "ex":
		return value * .43
	case "pt":
		return value / 10
	case "pc":
		return value * 1.2
	case "px":
		return jscompat.NumberStep(value*7.2) / 72
	case "in":
		return value * 7.2
	case "cm":
		return jscompat.NumberStep(value*7.2) / 2.54
	case "mm":
		return jscompat.NumberStep(value*7.2) / 25.4
	case "mu":
		return value / 18
	}
	return 0
}

// Matrix consumes a braced stack item, or a single character followed by a
// synthetic close brace. It does not read an unbraced control sequence as a
// complete TeX argument.
func (p *parser) readMatrixBody(name string) (string, error) {
	if err := p.startMatrixBody(name); err != nil {
		return "", err
	}
	start, depth, environments := p.pos, 0, 0
	for p.pos < len(p.source) {
		switch p.consumeRune() {
		case '%':
			p.skipComment()
		case '\\':
			command := p.readControlSequence()
			if name == "cases" && depth == 0 && (command == "begin" || command == "end") {
				if _, _, err := p.readArgument(command, false); err != nil {
					return "", err
				}
				if command == "begin" {
					environments++
				} else {
					environments--
				}
			}
		case '&':
			if name == "cases" && depth == 0 && environments == 0 {
				// Delay text errors until its first cell has been parsed.
				// Entry runs only after the preceding math reaches its &.
				end, _, _ := matrixCasesTextEnd(p.source, p.pos)
				p.pos = end
			}
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return p.source[start : p.pos-1], nil
			}
			depth--
		}
	}
	return "", texError("MissingCloseBrace", "Missing close brace")
}

// Matrix reads its opening argument before pushing the open ArrayItem. Script
// parsing uses the same boundary to reject an unbraced ArrayItem immediately.
func (p *parser) startMatrixBody(name string) error {
	p.skipNextSpaces()
	if p.pos == len(p.source) {
		return texError("MissingArgFor", "Missing argument for \\%s", name)
	}
	if p.source[p.pos] == '{' {
		p.pos++
	} else {
		c := p.consumeRune()
		p.source = string(c) + "}" + p.source[p.pos:]
		p.pos = 0
	}
	return nil
}

type matrixSourceRow struct {
	cells      []string
	spacing    string
	casesError error
}

func splitMatrixBody(body string, cases bool) ([]matrixSourceRow, error) {
	rows := []matrixSourceRow{{}}
	scanner := &parser{source: body}
	textSource := body
	if cases {
		// Entry sees the matrix's closing brace, including the non-letter
		// after a final \cr. Keep it available to the text scanner.
		textSource += "}"
	}
	start, depth, environments := 0, 0, 0
	for scanner.pos < len(body) {
		at := scanner.pos
		switch scanner.consumeRune() {
		case '%':
			scanner.skipComment()
		case '{':
			depth++
		case '}':
			depth--
		case '&':
			if depth == 0 && environments == 0 {
				last := &rows[len(rows)-1]
				last.cells = append(last.cells, body[start:at])
				start = scanner.pos
				if cases {
					end, _, err := matrixCasesTextEnd(textSource, scanner.pos)
					if err != nil {
						last.casesError = err
					}
					scanner.pos = end
				}
			}
		case '\\':
			command := scanner.readControlSequence()
			if depth != 0 {
				continue
			}
			if command == "begin" || command == "end" {
				if _, _, err := scanner.readArgument(command, false); err != nil {
					return nil, err
				}
				if command == "begin" {
					environments++
				} else {
					environments--
				}
			}
			if environments != 0 || command != "\\" && command != "cr" {
				continue
			}
			last := &rows[len(rows)-1]
			last.cells = append(last.cells, body[start:at])
			if command == "\\" {
				if scanner.pos < len(body) && body[scanner.pos] == '*' {
					scanner.pos++
				}
				if scanner.pos < len(body) && body[scanner.pos] == '[' {
					dimension, _, err := scanner.readBrackets(nil)
					if err != nil {
						return nil, err
					}
					if dimension != "" {
						match := dimensionFull.FindStringSubmatch(dimension)
						if match == nil {
							return nil, texError("BracketMustBeDimension", "Bracket argument to \\\\ must be a dimension")
						}
						last.spacing = dimensionValue(match)
					}
				}
			}
			rows = append(rows, matrixSourceRow{})
			start = scanner.pos
		}
	}
	last := &rows[len(rows)-1]
	last.cells = append(last.cells, body[start:])
	return rows, nil
}

func (p *parser) parseMatrixCell(source string, terminator byte) (*mml.Node, error) {
	sub := p.matrixCellParser(source + string(terminator))
	// Unlike an ordinary OpenItem, Matrix's ArrayItem requires a close brace;
	// a Right/Middle closing item reaches that boundary through styles/Over.
	children, _, err := sub.parseRowWithInfix(terminator, false, false)
	if err != nil {
		return nil, err
	}
	return sub.arrayCell.finish(matrixCellContent(children), len(children)), nil
}

func (p *parser) matrixCellParser(source string) *parser {
	// ArrayItem resets its lexical environment at entry and after every cell.
	// Keep the configuration and logical Stack.global while starting without
	// the surrounding font, root-index, or identifier-pattern state.
	return &parser{source: source, state: p.state, stackGlobal: p.ensureStackGlobal(), display: p.display,
		liveMatrix:       p.liveMatrix,
		environmentOwner: p.environmentOwner, environmentRow: p.environmentOwner,
		vectorFactory: p.vectorFactory, genfracPalette: p.genfracPalette,
		starMacroChildren: p.starMacroChildren, derivativeChildren: p.derivativeChildren, matrixClose: true, arrayCell: &arrayCellState{}}
}

func matrixCellContent(children []*mml.Node) *mml.Node {
	content := row(children, true)
	content.Walk(func(n *mml.Node) bool {
		if n.Flags.Token {
			n.SetProperty(resolvedFontScope, true)
		}
		return true
	})
	return content
}
