// Copyright (c) 2019-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/physics/{PhysicsMethods,PhysicsMappings}.ts.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) matrixQuantity(name string) ([]*mml.Node, error) {
	// MatrixQuantity uses GetStar and GetNext's JavaScript whitespace rules.
	star := p.readMathtoolsStar()
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	array := "array"
	if name == "smqty" || name == "smallmatrixquantity" {
		array = "smallmatrix"
	}
	arg, open, close := "", "", ""
	var err error
	next := byte(0)
	if p.pos < len(p.source) {
		next = p.source[p.pos]
	}
	switch next {
	case '{':
		arg, _, err = p.readArgumentAtCursor(name, false)
	case '(', '[', '|':
		p.pos++
		open, close = string(next), string(next)
		if next == '(' {
			close = ")"
			if star {
				open, close = "\\lgroup", "\\rgroup"
			}
		} else if next == '[' {
			close = "]"
		}
		terminator := string(next)
		if next == '(' {
			terminator = ")"
		} else if next == '[' {
			terminator = "]"
		}
		arg, err = p.readUpTo(name, terminator)
	default:
		// The fallback renders an empty fenced array without consuming the
		// next token. Only the GetNext whitespace has been consumed.
		open, close = "(", ")"
	}
	if err != nil {
		return nil, err
	}
	// In array the {} supplies the column definition; in smallmatrix it
	// remains a real empty group in the first cell. A fresh child parser
	// preserves both outcomes and the registered environment/fence commands.
	expansion := "\\begin{" + array + "}{} " + arg + "\\end{" + array + "}"
	if open != "" {
		expansion = "\\left" + open + expansion + "\\right" + close
	}
	parsed, err := p.parseChild(expansion)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(parsed), nil
}
