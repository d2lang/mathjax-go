// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/mathtools/MathtoolsMethods.ts (MtLap), ParseUtil.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) textLap(name string) ([]*mml.Node, error) {
	// GetArgument starts with GetNext's JavaScript whitespace. This boundary
	// is separate from the whitespace interpreted inside the captured text.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	raw, _, err := p.readArgumentAtCursor(name, false)
	if err != nil {
		return nil, err
	}
	content, err := p.internalMath(raw, "", true)
	if err != nil {
		return nil, err
	}
	padded := setAttributes(node("mpadded", content...), map[string]any{"width": 0})
	switch name {
	case "textllap":
		padded.Attributes.Set("lspace", "-1width")
	case "clap", "textclap":
		padded.Attributes.Set("lspace", "-.5width")
	}
	// MtLap pushes the padded box itself, unlike the mathematical Lap
	// handlers, which add a TeXAtom.
	return []*mml.Node{padded}, nil
}
