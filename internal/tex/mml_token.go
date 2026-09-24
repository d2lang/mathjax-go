// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/base/BaseMethods.ts (MmlToken/MmlTokenAllow),
// ts/input/tex/ParseUtil.ts (MmlFilterAttribute), and NodeUtil.ts.

package tex

import (
	"errors"
	"regexp"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// JavaScript's non-Unicode /i [a-z] is ASCII, while its \s includes these
// Unicode whitespace characters. Go's regexp shorthand classes differ.
const mmlTokenSpace = `[\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`

var (
	mmlTokenLeadingSpace = regexp.MustCompile(`^` + mmlTokenSpace + `+`)
	mmlTokenAttribute    = regexp.MustCompile(`^([a-zA-Z]+)` + mmlTokenSpace + `*=` + mmlTokenSpace + `*('[^']*'|"[^"]*"|[^ ,]*)` + mmlTokenSpace + `*,?` + mmlTokenSpace + `*`)
)

func mmlTokenAllowed(name string) bool {
	switch name {
	case "fontfamily", "fontsize", "fontweight", "fontstyle", "color", "background", "id", "class", "href", "style":
		return true
	}
	return false
}

func (p *parser) mmlToken(name string) ([]*mml.Node, error) {
	kind, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	attributes, _, err := p.readBrackets(nil)
	if err != nil {
		var failure *Error
		if errors.As(err, &failure) && failure.ID == "MissingCloseBracket" {
			return nil, texError(failure.ID, "Could not find closing ']' for argument to %s", "\\"+name)
		}
		return nil, err
	}
	attributes = mmlTokenLeadingSpace.ReplaceAllString(attributes, "")
	text, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	definition, ok := texMMLFactory.Definition(kind)
	if !ok || !definition.Flags.Token {
		return nil, texError("NotMathMLToken", "%s is not a token element", kind)
	}
	n := p.noteMO(node(kind))
	var properties mjSourceObject
	var keep []string
	for attributes != "" {
		match := mmlTokenAttribute.FindStringSubmatch(attributes)
		if match == nil {
			return nil, texError("InvalidMathMLAttr", "Invalid MathML attribute: %s", attributes)
		}
		key := match[1]
		if !n.Attributes.HasDefault(key) && !mmlTokenAllowed(key) {
			return nil, texError("UnknownAttrForElement", "%s is not a recognized attribute for %s", key, kind)
		}
		value := match[2]
		// The pinned unquote regexp uses '.', without dotAll. Quoted values
		// containing a JS line terminator therefore retain their quotes.
		if len(value) >= 2 && (value[0] == '\'' || value[0] == '"') && value[len(value)-1] == value[0] && !strings.ContainsAny(value[1:len(value)-1], "\n\r\u2028\u2029") {
			value = value[1 : len(value)-1]
		}
		// MmlFilterAttribute is the identity function in the pinned package.
		if value != "" {
			var parsed any = value
			if strings.ToLower(value) == "true" {
				parsed = true
			} else if strings.ToLower(value) == "false" {
				parsed = false
			}
			properties = append(properties, mjSourceProperty{Name: key, Value: parsed})
			keep = append(keep, key)
		}
		attributes = attributes[len(match[0]):]
	}
	if len(keep) != 0 {
		properties = append(properties, mjSourceProperty{Name: "mjx-keep-attrs", Value: strings.Join(keep, " ")})
	}
	// Upstream's inherited-attribute pass trims MmlMspace's text child to
	// its zero arity. Do not retain that child in the compiled token.
	if definition.Flags.Arity != 0 {
		n.SetChildren([]*mml.Node{mml.NewText(text)})
	}
	applySourceObject(n, properties)
	return []*mml.Node{n}, nil
}
