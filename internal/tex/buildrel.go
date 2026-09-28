// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/base/BaseMethods.ts, TexParser.ts, FilterUtil.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) buildRelation(name string) ([]*mml.Node, error) {
	raw, err := p.readUpTo(name, "\\over")
	if err != nil {
		return nil, err
	}
	top, err := p.parseChild(raw)
	if err != nil {
		return nil, err
	}
	// ParseArg uses GetNext's JavaScript whitespace before capturing its
	// one argument. Keep this boundary independent of ordinary tokenization.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	baseSource, _, err := p.readArgumentAtCursor(name, false)
	if err != nil {
		return nil, err
	}
	base, err := p.parseChild(baseSource)
	if err != nil {
		return nil, err
	}
	// BuildRel creates munderover(base, null, top). FilterUtil.cleanSubSup
	// reduces that absent under-script to mover; construct the same final
	// shape, without Overset's accent or movable-limit adjustments.
	return []*mml.Node{texAtom(node("mover", base, top), mml.TeXClassRel)}, nil
}

// readUpTo ports TexParser.GetUpTo's lexical control-sequence search. Braces
// protect the terminator; macros and comments are not expanded or interpreted
// until ParseUpTo creates the independent child parser for the captured text.
func (p *parser) readUpTo(name, token string) (string, error) {
	skip := func() {
		for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
			p.consumeRune()
		}
	}
	skip()
	start, depth := p.pos, 0
	for p.pos < len(p.source) {
		end := p.pos
		skip()
		if p.pos >= len(p.source) {
			break
		}
		c := string(p.consumeRune())
		switch c {
		case "\\":
			c += p.readControlSequence()
		case "{":
			depth++
		case "}":
			if depth == 0 {
				return "", texError("ExtraCloseLooking", "Extra close brace while looking for %s", token)
			}
			depth--
		}
		if depth == 0 && c == token {
			return p.source[start:end], nil
		}
	}
	return "", texError("TokenNotFoundForCommand", "Could not find %s for \\%s", token, name)
}
