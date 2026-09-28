// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: PhysicsMethods.Quantity, before argument or AutoOpen construction.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Quantity peeks for an optional size command and an allowed fence. An
// unsupported control sequence emits empty fences and is left in the caller,
// even for a registration which otherwise requires a braced argument.
func (p *parser) quantityFallback(name, open, close string) ([]*mml.Node, bool, error) {
	if name == "norm" {
		// ParseUtil.fenced receives the registration's literal strings;
		// this empty fallback does not parse the TeX delimiter command.
		open, close = "\\|", "\\|"
	}
	argument := name != "qty" && name != "quantity"
	start := p.pos
	getNext := func() rune {
		for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
			p.consumeRune()
		}
		return p.peekRune()
	}
	next := getNext()
	if argument && next == '*' {
		p.pos++
		next = getNext()
	}
	position := p.pos
	empty := func() ([]*mml.Node, bool, error) {
		p.pos = position
		return []*mml.Node{p.leftRightFenced(open, forcedRow(nil, false), close, true)}, true, nil
	}
	if next == '\\' {
		p.pos++
		switch p.readControlSequence() {
		case "big", "Big", "bigg", "Bigg":
			next = getNext()
		default:
			return empty()
		}
	}
	if argument && next != '{' {
		return nil, true, texError("MissingArgFor", "Missing argument for \\"+name)
	}
	switch next {
	case '(', '[', '{', '|':
		// The ordinary argument/AutoOpen handler owns supported fences.
		p.pos = start
		return nil, false, nil
	default:
		return empty()
	}
}
