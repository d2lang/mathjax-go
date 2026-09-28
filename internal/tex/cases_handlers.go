// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation of cases/CasesConfiguration.ts and EmpheqUtil.left.
package tex

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// NumCases is an EqnArray owner. Its left program is retained verbatim until
// the array closes, so declarations made in either rows or that program can
// resolve the appended helper through the ordinary shared command maps.
func (p *parser) casesEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if name != "numcases" && name != "subnumcases" {
		return nil, false, nil
	}
	left, _, err := p.readArgument("begin", false)
	if err != nil {
		return nil, true, err
	}
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, true, err
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}

	tags := p.amsTags()
	tags.start(name, true, true)
	ended := false
	defer func() {
		if !ended {
			tags.end()
		}
	}()
	table := node("mtable")
	maximum := 0
	appendRow := func(row *mml.Node) error {
		if len(row.Children) > maximum {
			maximum = len(row.Children)
		}
		tag, err := tags.getTag(p)
		if err != nil {
			return err
		}
		if tag != nil {
			row = node("mlabeledtr", append([]*mml.Node{tag}, row.Children...)...)
		}
		table.AppendChild(row)
		tags.clearTag()
		return nil
	}
	spacing := newEquationTableState(appendRow)
	spacing.initialSpacing = ".2em"
	row := &equationRowState{table: spacing}
	err = p.parseArrayBody(body, arrayBodyOwner{
		configure: func(sub *parser) {
			sub.arrayCell.equation, sub.arrayCell.numCases = row, true
		},
		prepareEntry: func(sub *parser, item *cellItem) ([]*mml.Node, error) {
			if item == nil || !item.numCasesText {
				return nil, nil
			}
			return sub.prepareNumCasesText(name)
		},
		hasEntries: func() bool { return len(row.entries) != 0 },
		endEntry: func(children []*mml.Node, _ *arrayCellState) error {
			row.endEntry(children)
			return nil
		},
		endRow:     row.endRow,
		addSpacing: spacing.addSpacing,
	})
	if err != nil {
		return nil, true, err
	}
	resetTableAttributes(table,
		"displaystyle", false,
		"rowspacing", ".2em",
		"columnalign", repeatAMSEqnArrayDefinition("left left", maximum),
		"columnspacing", "1em",
		"side", "right",
		"minlabelspacing", "0.8em",
	)
	spacing.applySpacing(table)
	// EqnArray EndTable restores the tag stack before Empheq parses left.
	tags.end()
	ended = true
	original := p.copyNode(table)
	if err := p.empheqAddLeft(table, original, left+"\\empheqlbrace\\,", "numcases-left"); err != nil {
		return nil, true, err
	}
	return []*mml.Node{table}, true, nil
}

// Cases.Entry's scanner differs from BaseMethods.Entry: only leading JS
// whitespace is trimmed; \label ends the scan; a complete \text remains raw
// input to internalMath; commands inside braces do not end the scan.
func numCasesTextEnd(source string, start int) (int, error) {
	braces := 0
	for i := start; i < len(source); {
		switch source[i] {
		case '{':
			braces++
			i++
		case '}':
			if braces == 0 {
				return i, nil
			}
			braces--
			i++
		case '&':
			if braces == 0 {
				return i, texError("ExtraCasesAlignTab", "Extra alignment tab in text for numcase environment")
			}
			i++
		case '\\':
			if braces != 0 {
				i++
				continue
			}
			if i+1 == len(source) {
				return i, fmt.Errorf("NumCases text scanner has no control-sequence token")
			}
			j := i + 1
			for j < len(source) && isASCIILetter(rune(source[j])) {
				j++
			}
			if j == i+1 {
				r, size := utf8.DecodeRuneInString(source[j:])
				if r == '\n' || r == '\r' || r == '\u2028' || r == '\u2029' {
					return i, fmt.Errorf("NumCases text scanner has no control-sequence token")
				}
				j += size
			}
			command := source[i+1 : j]
			if command == "\\" || command == "cr" || command == "end" || command == "label" {
				return i, nil
			}
			// The source advances by cs.length from the backslash position,
			// then examines the last character on its next iteration. In
			// particular, a nonletter brace is still a brace to this scanner.
			if isASCIILetter(rune(source[i+1])) {
				i = j - 1
			} else {
				i++
			}
		default:
			_, size := utf8.DecodeRuneInString(source[i:])
			i += size
		}
	}
	return len(source), nil
}

func (p *parser) prepareNumCasesText(environment string) ([]*mml.Node, error) {
	// The captured program is followed by this actual closing environment in
	// the original parser. It is visible to the raw scanner, not body math.
	end, err := numCasesTextEnd(p.source+"\\end{"+environment+"}", p.pos)
	if err != nil {
		return nil, err
	}
	if end > len(p.source) {
		end = len(p.source)
	}
	text := strings.TrimLeftFunc(p.source[p.pos:end], internalTextSpace)
	p.pos = end
	return p.internalMath(text, "", true)
}
