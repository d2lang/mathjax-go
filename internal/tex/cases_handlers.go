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

// NumCases pushes the same CasesBegin object twice below its EqnArray.
// The first matching End removes one entry; left decoration runs on Last
// (the oldest node), and the second End removes the remaining entry.
type casesBeginItem struct {
	end     bool
	entries int
	closing *environmentEndItem
}

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
	// EqnArray first pushes CasesBegin, reducing pending row items, before
	// checking equation nesting or parsing any physical body input.
	if p.commandCasesBegin != nil {
		if err := p.commandCasesBegin(); err != nil {
			return nil, true, err
		}
	}
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, true, err
	}
	frame := p.environmentOwner
	frame.casesBegin = &casesBeginItem{end: true, entries: 2}
	array := &ordinaryArrayItem{begin: frame}

	tags := p.amsTags()
	tags.start(name, true, true)
	ended := false
	defer func() {
		if !ended && !array.popped {
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
	equation := &equationRowState{table: spacing}
	err = p.parseArrayBody("", arrayBodyOwner{
		environment: frame,
		ordinary:    array,
		rules:       spacing.rules,
		configure: func(sub *parser) {
			sub.arrayCell.equation, sub.arrayCell.numCases = equation, true
		},
		prepareEntry: func(sub *parser, item *cellItem) ([]*mml.Node, error) {
			if item == nil || !item.numCasesText {
				return nil, nil
			}
			return sub.prepareNumCasesText(name)
		},
		hasEntries: func() bool { return len(equation.entries) != 0 },
		endEntry: func(children []*mml.Node, _ *arrayCellState) error {
			equation.endEntry(children)
			return nil
		},
		endRow:     equation.endRow,
		addSpacing: spacing.addSpacing,
	})
	if err != nil {
		return nil, true, err
	}
	if array.popped {
		// Stack.Pop().toMml() delivers only pending nodes to CasesBegin.
		// Completed equation rows and EndTable/tag cleanup are bypassed.
		children, err := p.continueCasesBegin(frame, []*mml.Node{row(array.nodes, true)})
		if err != nil {
			return nil, true, err
		}
		if frame.closed && frame.casesBegin.closing == nil {
			// Both aliases were directly popped. Their final MML goes to
			// the outer recipient without running the Cases End handler.
			return children, true, nil
		}
		if frame.closed {
			// The first matching End discarded the final alias with no
			// MML replacement. Select Last from the exposed recipient,
			// then replay the synthetic second End on that same row.
			p.commandCasesLeft = &environmentEndItem{name: frame.casesBegin.closing.name}
			return nil, true, nil
		}
		if len(children) == 0 {
			return nil, true, fmt.Errorf("Cases left block requires a pending node")
		}
		// NodeStack.Last is nodes[0], even after more input appends nodes.
		first := children[0]
		original := p.copyNode(first)
		if err := p.empheqAddLeft(first, original, left+"\\empheqlbrace\\,", "numcases-left"); err != nil {
			return nil, true, err
		}
		if err := p.closeEnvironment(frame, &environmentEndItem{name: frame.casesBegin.closing.name}); err != nil {
			return nil, true, err
		}
		return children, true, nil
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
	if array.end == nil {
		return nil, true, frame.missing()
	}
	if err := p.closeEnvironment(frame, array.end); err != nil {
		return nil, true, err
	}
	// Cases receives the finalized ArrayItem MML before Empheq decorates it.
	table = finishArrayRules(table)
	original := p.copyNode(table)
	if err := p.empheqAddLeft(table, original, left+"\\empheqlbrace\\,", "numcases-left"); err != nil {
		return nil, true, err
	}
	if err := p.closeEnvironment(frame, &environmentEndItem{name: frame.casesBegin.closing.name}); err != nil {
		return nil, true, err
	}
	return []*mml.Node{table}, true, nil
}

// Continue the same physical input in CasesBegin's lexical environment.
// Its shared first End must run before the deferred left child is parsed.
func (p *parser) continueCasesBegin(begin *environmentFrame, prefix []*mml.Node) ([]*mml.Node, error) {
	previous, array := p.environmentRow, p.ordinaryArray
	cell, matrixClose, entry, braket := p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner
	p.environmentRow, p.ordinaryArray = begin, nil
	p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = nil, false, false, nil
	defer func() {
		p.environmentRow, p.ordinaryArray = previous, array
		p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = cell, matrixClose, entry, braket
	}()
	children, _, err := p.parseRowContinuation(0, false, false, nil, "", prefix)
	if err != nil {
		return nil, err
	}
	if !begin.closed && (begin.casesBegin.end || begin.casesBegin.closing == nil) {
		return nil, begin.missing()
	}
	return children, nil
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
	// The actual end remains visible in this same-input program. The text
	// scanner stops before it; ordinary command dispatch owns the token.
	end, err := numCasesTextEnd(p.source, p.pos)
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

// Finish the deferred Cases handler on the actual exposed row recipient.
// An absent left property concatenates as JavaScript undefined, rather than
// reusing the removed CasesBegin's left program. Copy precedes mutation.
func (p *parser) finishCasesLeft(first *mml.Node, left string) error {
	if first == nil {
		return fmt.Errorf("Cases left block requires a pending node")
	}
	original := p.copyNode(first)
	return p.empheqAddLeft(first, original, left+"\\empheqlbrace\\,", "numcases-left")
}
