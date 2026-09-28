// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: PhysicsMappings.Bqty and PhysicsMethods.Quantity.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Bqty's registration requires a braced argument. Its empty fallback receives
// literal delimiter strings, while its supported argument parses those same
// strings as TeX commands in a fresh child parser.
func (p *parser) braceQuantity(name string) ([]*mml.Node, error) {
	const open, close = "\\{", "\\}"
	if result, handled, err := p.quantityFallback(name, open, close); handled {
		return result, err
	}
	getNext := func() rune {
		for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
			p.consumeRune()
		}
		return p.peekRune()
	}
	next := getNext()
	star := next == '*'
	if star {
		p.pos++
		next = getNext()
	}
	big := ""
	if next == '\\' {
		p.pos++
		big = p.readControlSequence()
		getNext()
	}
	// The fallback already validated the source GetNext/required-brace test.
	argument, _, err := p.readArgumentAtCursor(name, false)
	if err != nil {
		return nil, err
	}
	source := open + " " + argument + " " + close
	if !star {
		if big != "" {
			source = "\\" + big + "l" + open + " " + argument + " \\" + big + "r" + close
		} else {
			source = "\\left" + open + " " + argument + " \\right" + close
		}
	}
	result, err := p.parseChild(source)
	if err != nil {
		return nil, err
	}
	return []*mml.Node{result}, nil
}
