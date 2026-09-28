// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/base/BaseMethods.ts (FrameBox), TexParser.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) frameBox(name string) ([]*mml.Node, error) {
	var options [2]string
	for i := range options {
		value, _, err := p.readBrackets(nil)
		if err != nil {
			if failure, ok := err.(*Error); ok && failure.ID == "MissingCloseBracket" {
				return nil, texError(failure.ID, "Could not find closing ']' for argument to \\%s", name)
			}
			return nil, err
		}
		options[i] = value
	}
	// GetArgument's GetNext uses JavaScript whitespace, including BOM and
	// excluding NEL. Keep that local helper boundary separate from math text.
	for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
		p.consumeRune()
	}
	raw, _, err := p.readArgumentAtCursor(name, false)
	if err != nil {
		return nil, err
	}
	content, err := p.internalMath(raw, "", false)
	if err != nil {
		return nil, err
	}
	if options[0] != "" {
		align := "center"
		if options[1] == "l" {
			align = "left"
		} else if options[1] == "r" {
			align = "right"
		}
		padded := node("mpadded", content...)
		padded.Attributes.Set("width", options[0])
		padded.Attributes.Set("data-align", align)
		content = []*mml.Node{padded}
	}
	box := setAttributes(node("menclose", content...), map[string]any{"notation": "box"})
	return []*mml.Node{texAtom(box, mml.TeXClassOrd)}, nil
}
