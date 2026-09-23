// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods and BaseItems.SubsupItem (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (a *scriptAttachment) missingOpen() error {
	if a.marker == '_' {
		return texError("MissingOpenForSub", "Missing open brace for subscript")
	}
	return texError("MissingOpenForSup", "Missing open brace for superscript")
}

// A SubsupItem stays pending across commands that emit no stack item. In
// particular, a macro expansion and SetFont do not supply an empty argument.
// Keep this event boundary local to scripts; GetArgument callers are separate.
func (p *parser) parseScriptArgument(attachment *scriptAttachment, font string) (script *mml.Node, currentFont string, err error) {
	currentFont = font
	defer func() {
		if err == nil && currentFont != "" {
			applyScopedMathVariant(script, currentFont)
		}
	}()
	for {
		p.skipSpaces()
		if p.pos >= len(p.source) {
			return nil, currentFont, texError("MissingScript", "Missing superscript or subscript argument")
		}
		if p.source[p.pos] == '%' {
			p.skipComment()
			continue
		}
		if p.source[p.pos] == '}' {
			p.pos++
			return nil, currentFont, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
		}
		if p.source[p.pos] == '{' {
			p.pos++
			children, _, parseErr := p.parseRow('}', false)
			if parseErr != nil {
				return nil, currentFont, parseErr
			}
			return texAtom(row(children, true), mml.TeXClassOrd), currentFont, nil
		}
		if isPrimeRune(p.peekRune()) {
			p.consumeRune()
			if _, primeErr := p.startPrime(attachment.pendingBase()); primeErr != nil {
				return nil, currentFont, primeErr
			}
			return nil, currentFont, attachment.missingOpen()
		}
		if c := p.source[p.pos]; c == '^' || c == '_' {
			p.pos++
			p.skipSpaces()
			base := attachment.pendingBase()
			moves, _ := base.Property("movesupsub")
			if _, prepareErr := prepareScriptAttachment(base, c, limitsTruthy(moves)); prepareErr != nil {
				return nil, currentFont, prepareErr
			}
			return nil, currentFont, attachment.missingOpen()
		}
		if p.source[p.pos] == '\\' {
			start := p.pos
			p.pos++
			name := p.readControlSequence()
			if _, macro := p.state.macros[name]; !macro {
				if variant, ok := fontDeclarations[name]; ok {
					p.activeFont, currentFont = variant, variant
					continue
				}
				if _, ok := styleDeclarations[name]; ok {
					return nil, currentFont, attachment.missingOpen()
				}
				if _, ok := sizeDeclarations[name]; ok {
					return nil, currentFont, attachment.missingOpen()
				}
				if name == "right" {
					if _, parseErr := p.readDelimiter(false); parseErr != nil {
						return nil, currentFont, parseErr
					}
					return nil, currentFont, texError("MissingLeftExtraRight", "Missing \\left or extra \\right")
				}
			}
			p.pos = start
		}
		created, parseErr := p.parseOneToken()
		if parseErr != nil {
			return nil, currentFont, parseErr
		}
		if len(created) == 0 {
			continue
		}
		return row(created, true), currentFont, nil
	}
}
