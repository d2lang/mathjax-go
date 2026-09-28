// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/{TexParser,ParseUtil}.ts, ams/AmsMethods.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Genfrac needs GetDelimiterArg's raw key, not the generic converter's glyph.
// A nil key is absent; a present dot still constructs a zero-content fence.
func (p *parser) amsGenfracDelimiter(name string) (*string, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	key := strings.TrimFunc(raw, internalTextSpace)
	if strings.HasSuffix(key, "\\") && strings.HasSuffix(raw, " ") {
		key += " "
	}
	if key == "" {
		return nil, nil
	}
	if _, ok := lookupDelimiter(key); !ok {
		return nil, texError("MissingOrUnrecognizedDelim", "Missing or unrecognized delimiter for \\%s", name)
	}
	return &key, nil
}

// This caller-specific fixedFence path leaves binomial and infix helpers intact.
func (p *parser) amsGenfracFixedFence(open *string, fraction *mml.Node, close *string) (*mml.Node, error) {
	var openValue, closeValue any
	if open != nil {
		openValue = *open
	}
	if close != nil {
		closeValue = *close
	}
	result := forcedRow(nil, false)
	result.SetProperty("open", openValue)
	result.SetProperty("close", closeValue)
	result.SetProperty("texClass", mml.TeXClassOrd)
	result.TeXClass = mml.TeXClassOrd
	if open != nil {
		left, err := p.amsGenfracPalette(*open, "l")
		if err != nil {
			return nil, err
		}
		result.AppendChild(left)
	}
	result.AppendChild(fraction)
	if close != nil {
		right, err := p.amsGenfracPalette(*close, "r")
		if err != nil {
			return nil, err
		}
		result.AppendChild(right)
	}
	return result, nil
}

func (p *parser) amsGenfracPalette(fence, side string) (*mml.Node, error) {
	if fence == "{" || fence == "}" {
		fence = "\\" + fence
	}
	display := "{\\bigg" + side + " " + fence + "}"
	text := "{\\big" + side + " " + fence + "}"
	source := "\\mathchoice" + display + text + text + text
	// mathPalette creates an empty-environment TexParser with the same
	// configuration. Preserve shared tables/factory and the caller's count.
	count := p.state.macroCount
	p.state.macroCount = 0
	defer func() { p.state.macroCount = count }()
	sub := &parser{source: source, state: p.state, vectorFactory: p.vectorFactory,
		genfracPalette: true, derivativeChildren: p.derivativeChildren}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	// The source's empty lexical environment must survive a deferred outer
	// font walk, including tokens that selected their default variant.
	result.Walk(func(n *mml.Node) bool {
		if n.Flags.Token {
			n.SetProperty(resolvedFontScope, true)
		}
		return true
	})
	return result, nil
}

// Both Sqrt and Root create a fresh parseRoot index parser. Its counter is
// independent even when the surrounding expression is ordinary authored TeX.
func (p *parser) parseRootIndex(source string) (*mml.Node, error) {
	count := p.state.macroCount
	p.state.macroCount = 0
	defer func() { p.state.macroCount = count }()
	inRoot := p.inRoot
	p.inRoot = true
	defer func() { p.inRoot = inRoot }()
	global := &parserStackGlobal{}
	index, err := p.parseStringWithStack(source, global)
	if err != nil {
		return nil, err
	}
	if global.leftRoot == "" && global.upRoot == "" {
		return index, nil
	}
	padded := node("mpadded", index)
	if global.leftRoot != "" {
		padded.Attributes.Set("width", global.leftRoot)
	}
	if global.upRoot != "" {
		padded.Attributes.Set("voffset", global.upRoot)
		padded.Attributes.Set("height", global.upRoot)
	}
	return padded, nil
}
