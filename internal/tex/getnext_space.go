// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation and modification of TexParser.GetNext/nextIsSpace.

package tex

// skipNextSpaces advances the original GetNext boundary. Its JavaScript \s
// includes BOM and excludes NEL, MVS, and zero-width space.
func (p *parser) skipNextSpaces() {
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
}
