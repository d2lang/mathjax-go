// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation of MathJax 3.2.2 ParseUtil.internalMath/internalText and HBox.

package tex

import (
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) hboxCommand(name, font string, levelZero bool) ([]*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	return p.internalMath(raw, font, levelZero)
}

// internalMath scans only the delimiters and escapes recognized by the source
// helper. Commands in literal text are not parsed as TeX, except ref/eqref.
func (p *parser) internalMath(text, font string, levelZero bool) ([]*mml.Node, error) {
	if font == "" {
		font = p.activeFont
	}
	var nodes []*mml.Node
	i, start, braces := 0, 0, 0
	var match byte
	appendText := func(end int) {
		if start < end {
			nodes = append(nodes, internalText(text[start:end], font))
		}
	}
	appendMath := func(end int, reference bool) error {
		inner, err := p.parseInternalMath(text[start:end])
		if err != nil {
			return err
		}
		atom := texAtom(inner, mml.TeXClassOrd)
		if reference && font != "" {
			atom.Attributes.Set("mathvariant", font)
		}
		nodes = append(nodes, atom)
		return nil
	}
	for i < len(text) {
		c := text[i]
		i++
		switch c {
		case '$':
			if match == '$' && braces == 0 {
				if err := appendMath(i-1, false); err != nil {
					return nil, err
				}
				match, start = 0, i
			} else if match == 0 {
				appendText(i - 1)
				match, start = '$', i
			}
		case '{':
			if match != 0 {
				braces++
			}
		case '}':
			if match == '}' && braces == 0 {
				if err := appendMath(i, true); err != nil {
					return nil, err
				}
				match, start = 0, i
			} else if match != 0 && braces > 0 {
				braces--
			}
		case '\\':
			if length := internalReferencePrefix(text[i:]); match == 0 && length > 0 {
				appendText(i - 1)
				match, start = '}', i-1
				i += length
				continue
			}
			if i == len(text) {
				continue
			}
			c = text[i]
			i++
			if c == '(' && match == 0 {
				appendText(i - 2)
				match, start = ')', i
			} else if c == ')' && match == ')' && braces == 0 {
				if err := appendMath(i-2, false); err != nil {
					return nil, err
				}
				match, start = 0, i
			} else if match == 0 && strings.ContainsRune("${}\\", rune(c)) {
				i--
				text = text[:i-1] + text[i:]
			}
		}
	}
	if match != 0 {
		return nil, texError("MathNotTerminated", "Math not terminated in text box")
	}
	appendText(len(text))
	if levelZero {
		style := node("mstyle", nodes...)
		style.Attributes.Set("displaystyle", false)
		style.Attributes.Set("scriptlevel", 0)
		return []*mml.Node{style}, nil
	}
	if len(nodes) > 1 {
		return []*mml.Node{node("mrow", nodes...)}, nil
	}
	return nodes, nil
}

func internalText(text, font string) *mml.Node {
	if trimmed := strings.TrimLeftFunc(text, internalTextSpace); trimmed != text {
		text = "\u00a0" + trimmed
	}
	if trimmed := strings.TrimRightFunc(text, internalTextSpace); trimmed != text {
		text = trimmed + "\u00a0"
	}
	n := node("mtext", mml.NewText(text))
	if font != "" {
		n.Attributes.Set("mathvariant", font)
	}
	return n
}

func internalReferencePrefix(text string) int {
	index := 0
	if strings.HasPrefix(text, "eqref") {
		index = 5
	} else if strings.HasPrefix(text, "ref") {
		index = 3
	} else {
		return 0
	}
	for index < len(text) {
		r, size := utf8.DecodeRuneInString(text[index:])
		if !internalTextSpace(r) {
			break
		}
		index += size
	}
	if index < len(text) && text[index] == '{' {
		return index + 1
	}
	return 0
}

// JavaScript whitespace includes BOM and excludes NEL.
func internalTextSpace(r rune) bool {
	return r >= '\t' && r <= '\r' || r == ' ' || r == 0xa0 || r == 0x1680 ||
		r >= 0x2000 && r <= 0x200a || r == 0x2028 || r == 0x2029 ||
		r == 0x202f || r == 0x205f || r == 0x3000 || r == 0xfeff
}

func (p *parser) parseInternalMath(source string) (*mml.Node, error) {
	// The configuration (macros, tags, colors and token factory) is shared,
	// while the source's fresh TexParser has an empty lexical environment and
	// its own expansion count. Nested parsing is synchronous; restore the
	// caller's counter even when the inner parser reports an error.
	count := p.state.macroCount
	p.state.macroCount = 0
	defer func() { p.state.macroCount = count }()
	sub := &parser{source: source, state: p.state, vectorFactory: p.vectorFactory}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	// A deferred enclosing font walk must respect the choices already made
	// in this empty environment, including a token's default variant.
	result.Walk(func(n *mml.Node) bool {
		if n.Flags.Token {
			n.SetProperty(resolvedFontScope, true)
		}
		return true
	})
	return result, nil
}
