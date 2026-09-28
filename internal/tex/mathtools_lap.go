// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/mathtools/MathtoolsMethods.ts (MathLap), TexParser.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) mathLap(name string) ([]*mml.Node, error) {
	style, _, err := p.readBrackets(nil)
	if err != nil {
		if failure, ok := err.(*Error); ok && failure.ID == "MissingCloseBracket" {
			return nil, texError(failure.ID, "Could not find closing ']' for argument to \\%s", name)
		}
		return nil, err
	}
	// ParseArg captures one argument after GetNext's JavaScript whitespace,
	// then parses it in a child TexParser with its own macro counter.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	raw, _, err := p.readArgumentAtCursor(name, false)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseChild(raw)
	if err != nil {
		return nil, err
	}
	padded := setAttributes(node("mpadded", arg), map[string]any{"width": 0})
	if strings.HasSuffix(name, "llap") {
		padded.Attributes.Set("lspace", "-1width")
	} else if strings.HasSuffix(name, "clap") {
		padded.Attributes.Set("lspace", "-.5width")
	}
	styled := node("mstyle", padded)
	styled.Attributes.Set("data-cramped", strings.HasPrefix(name, "cramped"))
	setMathtoolsDisplayLevel(styled, style)
	return []*mml.Node{texAtom(styled, mml.TeXClassOrd)}, nil
}
