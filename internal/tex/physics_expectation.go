// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: PhysicsMethods.Expectation and outputBraket.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) physicsExpectation(name string) ([]*mml.Node, error) {
	star1 := p.readStar()
	star2 := star1 && p.readStar()
	arg1, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	// Source GetNext commits whitespace even when no second argument follows.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	arg2 := ""
	if p.pos < len(p.source) && p.source[p.pos] == '{' {
		arg2, _, err = p.readArgument(name, true)
		if err != nil {
			return nil, err
		}
	}
	// Both raw arguments must be truthy. A consumed empty argument does not
	// select the three-operand expansion, and discarded source is not parsed.
	var program string
	if arg1 != "" && arg2 != "" {
		switch {
		case star2:
			program = "\\left\\langle{" + arg2 + "}\\middle\\vert{" + arg1 + "}\\middle\\vert{" + arg2 + "}\\right\\rangle"
		case star1:
			program = "\\langle{" + arg2 + "}\\vert{" + arg1 + "}\\vert{" + arg2 + "}\\rangle"
		default:
			program = "\\left\\langle{" + arg2 + "}\\right\\vert{" + arg1 + "}\\left\\vert{" + arg2 + "}\\right\\rangle"
		}
	} else if star1 {
		program = "\\langle {" + arg1 + "} \\rangle"
	} else {
		program = "\\left\\langle {" + arg1 + "} \\right\\rangle"
	}
	// One fresh TexParser copies the complete lexical environment. Both
	// occurrences of arg2 share its budget, while registrations remain shared.
	parsed, err := p.parseChild(program)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(parsed), nil
}
