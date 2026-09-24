// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on PhysicsMethods.Derivative/Commutator, AmsMethods, and TexParser.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Derivative, Commutator, SideSet and OperatorName create genuine child TexParsers.
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
