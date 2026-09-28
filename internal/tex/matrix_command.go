// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/base/{BaseMethods,BaseItems}.ts.

package tex

import (
	"math"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) matrixCommand(name string) ([]*mml.Node, error) {
	body, err := p.readMatrixBody(name)
	if err != nil {
		return nil, err
	}
	rows, err := splitMatrixBody(body)
	if err != nil {
		return nil, err
	}
	table := node("mtable")
	for i, sourceRow := range rows {
		cells := sourceRow.cells
		if i == len(rows)-1 && len(cells) == 1 && strings.TrimSpace(cells[0]) == "" {
			continue
		}
		line := node("mtr")
		for _, raw := range cells {
			content, err := p.parseMatrixCell(raw)
			if err != nil {
				return nil, err
			}
			line.AppendChild(node("mtd", content))
		}
		table.AppendChild(line)
	}
	table.Attributes.Set("rowspacing", "4pt")
	table.Attributes.Set("columnspacing", "1em")
	if name == "eqalign" {
		table.Attributes.Set("rowspacing", ".5em")
		table.Attributes.Set("columnspacing", "0.278em")
		table.Attributes.Set("displaystyle", true)
		table.Attributes.Set("columnalign", "right left")
	}
	for _, r := range rows {
		if r.spacing == "" {
			continue
		}
		base := .4
		if name == "eqalign" {
			base = .5
		}
		spacing := make([]string, len(table.Children))
		for i := range spacing {
			value := math.Max(0, base+matrixDimensionEm(rows[i].spacing))
			spacing[i] = "0em"
			if math.Abs(value) >= .0006 {
				spacing[i] = strings.TrimSuffix(strings.TrimRight(jscompat.ToFixed(value, 3), "0"), ".") + "em"
			}
		}
		table.Attributes.Set("rowspacing", strings.Join(spacing, " "))
		break
	}
	if name == "pmatrix" {
		return []*mml.Node{p.fenced("(", table, ")", true)}, nil
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
	p.skipSpaces()
	if p.pos == len(p.source) {
		return "", texError("MissingArgFor", "Missing argument for \\%s", name)
	}
	if p.source[p.pos] == '{' {
		p.pos++
	} else {
		c := p.consumeRune()
		p.source = string(c) + "}" + p.source[p.pos:]
		p.pos = 0
	}
	start, depth := p.pos, 0
	for p.pos < len(p.source) {
		switch p.consumeRune() {
		case '%':
			p.skipComment()
		case '\\':
			p.readControlSequence()
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

type matrixSourceRow struct {
	cells   []string
	spacing string
}

func splitMatrixBody(body string) ([]matrixSourceRow, error) {
	rows := []matrixSourceRow{{}}
	scanner := &parser{source: body}
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

func (p *parser) parseMatrixCell(source string) (*mml.Node, error) {
	// ArrayItem resets its lexical environment at entry and after every cell.
	// Keep the configuration and logical Stack.global while starting without
	// the surrounding font, root-index, or identifier-pattern state.
	sub := &parser{source: source + "}", state: p.state, stackGlobal: p.ensureStackGlobal(), display: p.display,
		vectorFactory: p.vectorFactory, genfracPalette: p.genfracPalette,
		starMacroChildren: p.starMacroChildren, derivativeChildren: p.derivativeChildren}
	children, _, err := sub.parseRow('}', false)
	if err != nil {
		return nil, err
	}
	content := row(children, true)
	content.Walk(func(n *mml.Node) bool {
		if n.Flags.Token {
			n.SetProperty(resolvedFontScope, true)
		}
		return true
	})
	return content, nil
}
