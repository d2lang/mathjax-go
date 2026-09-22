// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Prime and BaseItems.PrimeItem/SubsupItem (MathJax 3.2.2).
package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// pendingPrime belongs to a single parseRow stack. Keep the actual original
// node and prime token until a stack item finalizes or consumes them; an eager
// msup loses this distinction and wrongly treats x'^n as a double exponent.
type pendingPrime struct{ base, prime *mml.Node }

func isPrimeRune(r rune) bool { return r == '\'' || r == '\u2019' }

func (p *parser) startPrime(base *mml.Node) (*pendingPrime, error) {
	origin, _ := base.Property(limitsScriptOrigin)
	side := base.Kind == "msub" || base.Kind == "msup" || base.Kind == "msubsup"
	authoredSup := base.Kind == "msup" && origin != true
	if side && !authoredSup && ((base.Kind == "msup" && len(base.Children) > 1) || (base.Kind == "msubsup" && len(base.Children) > 2 && base.Children[2] != nil)) {
		return nil, texError("DoubleExponentPrime", "Prime causes double exponent: use braces to clarify")
	}
	count := 1
	for p.pos < len(p.source) {
		p.skipSpaces()
		if p.pos >= len(p.source) || !isPrimeRune(p.peekRune()) {
			break
		}
		p.consumeRune()
		count++
	}
	primes := []string{"", "′", "″", "‴", "⁗"}
	text := strings.Repeat("′", count)
	if count < len(primes) {
		text = primes[count]
	}
	prime := token("mo", text)
	prime.SetProperty("variantForm", true)
	return &pendingPrime{base, prime}, nil
}

func (p *pendingPrime) finish() *mml.Node {
	origin, _ := p.base.Property(limitsScriptOrigin)
	if p.base.Kind == "msub" || p.base.Kind == "msubsup" || (p.base.Kind == "msup" && origin == true) {
		var under *mml.Node
		if p.base.Kind != "msup" && len(p.base.Children) > 1 {
			under = p.base.Children[1]
		}
		if under != nil {
			if p.base.Kind != "msub" || origin == true {
				p.base.Kind = "msubsup"
			}
			p.base.Flags.Arity = 3
			p.base.SetChildren([]*mml.Node{p.base.Children[0], under, p.prime})
		} else {
			p.base.Kind = "msup"
			p.base.Flags.Arity = 2
			p.base.SetChildren([]*mml.Node{p.base.Children[0], p.prime})
		}
		refreshDynamicFlags(p.base)
		return p.base
	}
	return node("msup", p.base, p.prime)
}

func (p *pendingPrime) attach(script *mml.Node, marker byte) (*mml.Node, error) {
	moves, hasMoves := p.base.Property("movesupsub")
	if marker == '^' {
		p.prime.SetProperty("variantForm", true)
		script = node("mrow", p.prime, script)
	}
	result, err := attachScriptBase(p.base, script, marker, limitsTruthy(moves))
	if err != nil {
		return nil, err
	}
	if marker == '_' {
		result.Kind = map[bool]string{false: "msubsup", true: "munderover"}[result.Kind == "munder" || result.Kind == "mover" || result.Kind == "munderover"]
		result.Flags.Arity = 3
		result.SetChildren([]*mml.Node{result.Children[0], result.Children[1], p.prime})
		refreshDynamicFlags(result)
	}
	result.Flags.Embellished = p.base.Flags.Embellished
	result.Flags.CoreIndex = 0
	result.SetProperty(limitsScriptOrigin, true)
	if hasMoves {
		result.SetProperty("movesupsub", moves)
	}
	return result, nil
}
