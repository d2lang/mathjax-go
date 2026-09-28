// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on PhysicsMethods.Derivative/Differential/Commutator, AmsMethods, and TexParser.
package tex

import (
	"errors"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Derivative, Differential, Commutator, SideSet and OperatorName create child TexParsers.
// Configuration remains shared, but the lexical environment is copied and
// the parser's macro count starts at zero. Only source-proven child-parser
// boundaries inherit this local policy; sliced same-parser rows do not reset.
func (p *parser) parseChild(source string) (*mml.Node, error) {
	count := p.state.macroCount
	p.state.macroCount = 0
	defer func() { p.state.macroCount = count }()
	sub := &parser{source: source, state: p.state, display: p.display,
		multiLetterFont: p.multiLetterFont, activeFont: p.activeFont,
		identifierPattern: p.identifierPattern, operatorLetters: p.operatorLetters,
		noAutoOP: p.noAutoOP, fontExplicitEmpty: p.fontExplicitEmpty,
		vectorFactory: p.vectorFactory, vectorFont: p.vectorFont,
		vectorStar: p.vectorStar, vectorAlias: p.vectorAlias,
		genfracPalette: p.genfracPalette, starMacroChildren: p.starMacroChildren,
		derivativeChildren: true}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	if sub.activeFont != "" {
		applyScopedMathVariant(result, sub.activeFont)
	}
	return result, nil
}

// Differential reparses the operator, optional power and raw operand together.
// In particular, registered operators are honored and leading operand scripts
// attach to the differential rather than to an independently parsed empty base.
func (p *parser) differential(name string, after **derivativeAutoOpen) ([]*mml.Node, error) {
	power, hasPower, err := p.readBrackets(nil)
	if err != nil {
		var failure *Error
		if errors.As(err, &failure) && failure.ID == "MissingCloseBracket" {
			return nil, texError(failure.ID, "Could not find closing ']' for argument to %s", "\\"+name)
		}
		return nil, err
	}
	op := "\\diffd"
	if name == "variation" || name == "var" {
		op = "\\delta"
	}
	if hasPower {
		op += "^{" + power + "}"
	} else {
		op += " "
	}
	p.skipSpaces()
	parens := p.pos < len(p.source) && p.source[p.pos] == '('
	braces := p.pos < len(p.source) && p.source[p.pos] == '{'
	if !parens {
		arg, _, err := p.readArgument(name, !braces)
		if err != nil {
			return nil, err
		}
		op += arg
	}
	parsed, err := p.parseChild(op)
	if err != nil {
		return nil, err
	}
	if braces {
		return []*mml.Node{texAtom(parsed, mml.TeXClassOp)}, nil
	}
	if parens {
		*after = &derivativeAutoOpen{}
	}
	return unwrapInferred(parsed), nil
}
