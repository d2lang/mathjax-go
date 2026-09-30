// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods and BaseItems.SubsupItem (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Superscript and Subscript call GetNext once on entry and insert a space
// after an immediate ASCII digit, before checking the base's occupied slots.
// Keep this out of the pending argument loop: font/comment/macro continuation
// must still use the ordinary number scanner. Prime uses the same GetNext
// whitespace set; later token delivery uses the separate character-map spaces.
func (p *parser) scriptInitialLookahead() {
	for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
		p.consumeRune()
	}
	if p.pos < len(p.source) && p.source[p.pos] >= '0' && p.source[p.pos] <= '9' {
		p.source = p.source[:p.pos+1] + " " + p.source[p.pos+1:]
	}
}

func (a *scriptAttachment) missingOpen() error {
	if a.marker == '_' {
		return texError("MissingOpenForSub", "Missing open brace for subscript")
	}
	return texError("MissingOpenForSup", "Missing open brace for superscript")
}

// A SubsupItem stays pending across commands that emit no stack item. In
// particular, a macro expansion and SetFont do not supply an empty argument.
// Keep this event boundary local to scripts; GetArgument callers are separate.
func (p *parser) parseScriptArgument(attachment *scriptAttachment, font string) (script *mml.Node, currentFont string, trailing []*mml.Node, after *derivativeAutoOpen, err error) {
	// SubsupItem is the recipient, so an outer Braket cannot own this token.
	owner := p.braketOwner
	p.braketOwner = nil
	defer func() { p.braketOwner = owner }()
	currentFont = font
	defer func() {
		if err == nil && currentFont != "" {
			applyScopedMathVariant(script, currentFont)
		}
	}()
	for {
		p.skipMathTokenSpaces()
		if p.pos >= len(p.source) {
			return nil, currentFont, nil, nil, texError("MissingScript", "Missing superscript or subscript argument")
		}
		if p.source[p.pos] == '%' {
			p.skipComment()
			continue
		}
		if p.source[p.pos] == '}' {
			p.pos++
			return nil, currentFont, nil, nil, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
		}
		if p.source[p.pos] == '&' {
			return nil, currentFont, nil, nil, attachment.missingOpen()
		}
		if p.source[p.pos] == '{' {
			p.pos++
			children, _, parseErr := p.parseRow('}', false)
			if parseErr != nil {
				return nil, currentFont, nil, nil, parseErr
			}
			if p.environmentPopped {
				p.environmentPopped = false
				if len(children) == 0 {
					continue
				}
				return children[0], currentFont, children[1:], nil, nil
			}
			return texAtom(row(children, true), mml.TeXClassOrd), currentFont, nil, nil, nil
		}
		if isPrimeRune(p.peekRune()) {
			p.consumeRune()
			if _, primeErr := p.startPrime(attachment.pendingBase()); primeErr != nil {
				return nil, currentFont, nil, nil, primeErr
			}
			return nil, currentFont, nil, nil, attachment.missingOpen()
		}
		if c := p.source[p.pos]; c == '^' || c == '_' {
			p.pos++
			p.scriptInitialLookahead()
			base := attachment.pendingBase()
			moves, _ := base.Property("movesupsub")
			if _, prepareErr := prepareScriptAttachment(base, c, limitsTruthy(moves)); prepareErr != nil {
				return nil, currentFont, nil, nil, prepareErr
			}
			return nil, currentFont, nil, nil, attachment.missingOpen()
		}
		if p.source[p.pos] == '\\' {
			start := p.pos
			p.pos++
			name := p.readControlSequence()
			if _, macro := p.state.macros[name]; !macro {
				switch name {
				case "matrix", "array", "pmatrix", "cases", "eqalign", "eqalignno", "leqalignno", "displaylines":
					if name == "displaylines" {
						if _, paired := p.state.pairedDelimiters[name]; paired {
							break // The declared delimiter runs before Matrix.
						}
					}
					if parseErr := p.startMatrixBody(name); parseErr != nil {
						return nil, currentFont, nil, nil, parseErr
					}
					return nil, currentFont, nil, nil, attachment.missingOpen()
				}
				if variant, ok := fontDeclarations[name]; ok {
					p.activeFont, currentFont = variant, variant
					continue
				}
				if _, ok := styleDeclarations[name]; ok {
					return nil, currentFont, nil, nil, attachment.missingOpen()
				}
				if _, ok := sizeDeclarations[name]; ok {
					return nil, currentFont, nil, nil, attachment.missingOpen()
				}
				if name == "right" {
					if _, parseErr := p.readDelimiter(name, false); parseErr != nil {
						return nil, currentFont, nil, nil, parseErr
					}
					return nil, currentFont, nil, nil, texError("MissingLeftExtraRight", "Missing \\left or extra \\right")
				}
			}
			p.pos = start
		}
		result, parseErr := p.parseOneTokenEventWithCasesBegin(attachment.missingOpen)
		if parseErr != nil {
			return nil, currentFont, nil, nil, parseErr
		}
		if result.environmentEnd != nil {
			if !result.environmentEnd.spread {
				return nil, currentFont, nil, nil, result.environmentEnd.extra()
			}
			// SpreadLines pops this actual SubsupItem, whose nodes contain
			// the unfinished base, rather than feeding it a closing item.
			p.environmentPopped = true
			return p.publishPendingScript(attachment), currentFont, nil, nil, nil
		}
		if result.namedFunction || result.notItem || result.nonscriptItem || result.dotsItem != nil || result.positionItem != nil || result.cellItem != nil {
			return nil, currentFont, nil, nil, attachment.missingOpen()
		}
		if len(result.nodes) == 0 {
			// Empty PushAll leaves the SubsupItem pending. A following auto
			// open item is rejected here, before its tail can be parsed.
			if result.afterNode.start(p) {
				return nil, currentFont, nil, nil, attachment.missingOpen()
			}
			continue
		}
		// SubsupItem accepts the first final MML item. TexParser.PushAll
		// delivers any remaining items after that script has completed.
		return result.nodes[0], currentFont, result.nodes[1:], result.afterNode, nil
	}
}
