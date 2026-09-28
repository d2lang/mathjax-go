// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

// TexParser.Parse dispatches characters without a whitespace lookahead.
// Only these five BaseMappings entries call the no-op BaseMethods.Space.
// Keep token delivery distinct from GetNext's JavaScript whitespace set.
func isMathTokenSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n', '\u00a0':
		return true
	}
	return false
}

func (p *parser) skipMathTokenSpaces() {
	for p.pos < len(p.source) && isMathTokenSpace(p.peekRune()) {
		p.consumeRune()
	}
}
