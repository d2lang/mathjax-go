// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/physics/PhysicsMethods.ts, createVectorToken/VectorBold.
package tex

import (
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

const vectorFactoryToken = "go-vector-factory-token"
const vectorFactoryDone = "go-vector-factory-done"

// Token factory selection happens before an enclosing font scope is lowered.
// Each token is visited once in its creating parser's environment; subparser
// results must not be recolored when their parent later appends them.
func (p *parser) applyVectorFactory(n *mml.Node) {
	if !p.vectorFactory || n == nil {
		return
	}
	n.Walk(func(current *mml.Node) bool {
		origin, _ := current.Property(vectorFactoryToken)
		done, _ := current.Property(vectorFactoryDone)
		if origin != true || done == true {
			return true
		}
		current.SetProperty(vectorFactoryDone, true)
		text := textContent(current)
		r, size := utf8.DecodeRuneInString(text)
		// The source uses text.length === 1 (UTF-16), excluding supplementary runes.
		if p.activeFont != "" || p.vectorFont == "" || size == 0 || size != len(text) || r > 0xFFFF {
			return true
		}
		accent, _ := current.Attributes.Get("accent")
		selected := r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= 0x391 && r <= 0x3A9 || r >= '0' && r <= '9' || p.vectorStar && r >= 0x3B1 && r <= 0x3C9 || accent == true
		if selected {
			current.Attributes.Set("mathvariant", p.vectorFont)
		}
		return true
	})
}

func (p *parser) parseVectorString(source string, star bool) (*mml.Node, error) {
	if p.genfracPalette {
		// VectorBold creates a genuine child TexParser with its own count.
		count := p.state.macroCount
		p.state.macroCount = 0
		defer func() { p.state.macroCount = count }()
	}
	variant := "bold"
	if star {
		variant = "bold-italic"
	}
	sub := &parser{source: source, state: p.state, display: p.display,
		multiLetterFont: p.multiLetterFont, vectorFactory: true, vectorFont: variant, vectorStar: star,
		genfracPalette: p.genfracPalette}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	// Deleting the outer font applies to unselected and authored tokens too.
	// Protect their resulting default/explicit choice from deferred outer scopes.
	result.Walk(func(n *mml.Node) bool {
		if n.Flags.Token {
			n.SetProperty(resolvedFontScope, true)
		}
		return true
	})
	// The pinned handler deletes, rather than restores, the current vector env.
	// A nested VectorBold thus clears that env for later tokens in the same row.
	p.vectorFont = ""
	p.vectorStar = false
	return result, nil
}
