// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/{NodeFactory,ParseOptions,ParseUtil}.ts.

package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Registration belongs to this parse, including its argument subparsers. It
// records actual construction events before handlers can reorder or detach
// nodes. Pure constructors and the shared MML factory have no mutable hook.
func (p *parser) noteMO(n *mml.Node) *mml.Node {
	if n != nil && n.Kind == "mo" && p.state != nil {
		p.state.operators = append(p.state.operators, n)
	}
	return n
}

func (p *parser) token(kind, text string) *mml.Node {
	return p.noteMO(token(kind, text))
}

func (p *parser) operator(text string, class mml.TeXClass, attributes map[string]any) *mml.Node {
	return p.noteMO(operator(text, class, attributes))
}

// ParseUtil.copyNode registers the copied tree in preorder immediately after
// copying, including nodes later detached by pruning. It does not register
// nodes that have merely been reparented. Derivative instead reparses source
// through a child parser sharing this same parse state.
func (p *parser) copyNode(original *mml.Node) *mml.Node {
	copy := original.Clone()
	if copy != nil {
		copy.Walk(func(n *mml.Node) bool { p.noteMO(n); return true })
	}
	return copy
}

func (p *parser) namedOperator(text string) *mml.Node {
	return p.noteMO(namedOperator(text))
}

func (p *parser) lookupExtensionSymbol(name string) (*mml.Node, bool) {
	n, ok := lookupExtensionSymbol(name)
	return p.noteMO(n), ok
}

func (p *parser) cdHorizontalArrow(text string) *mml.Node {
	return p.noteMO(cdHorizontalArrow(text))
}

func (p *parser) cdVerticalArrow(text string) *mml.Node {
	return p.noteMO(cdVerticalArrow(text))
}

func (p *parser) amsFixedFencePalette(text string, class mml.TeXClass) *mml.Node {
	return amsFixedFencePaletteWithToken(text, class, p.token)
}

func (p *parser) fenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	return fencedWithOperator(open, content, close, stretchy, p.operator)
}

func (p *parser) leftRightFenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	n := p.fenced(open, content, close, stretchy)
	n.SetProperty("open", open)
	n.SetProperty("close", close)
	n.SetProperty("texClass", mml.TeXClassInner)
	n.TeXClass = mml.TeXClassInner
	return n
}

func (p *parser) physicsNabla() *mml.Node {
	return physicsNablaWithOperator(p.operator)
}

func (p *parser) empheqTopRowTable(original *mml.Node) *mml.Node {
	return empheqCopiedTopRowTable(p.copyNode(original))
}

// Source symbols create one token with no intervening parse. Register only that
// new node, retaining the source map's actual token kind and attributes.
func (p *parser) lookupMJSourceSymbol(name string) (*mml.Node, bool) {
	n, ok := lookupMJSourceSymbol(name)
	return p.noteMO(n), ok
}

func (p *parser) physicsFenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	n := p.fenced(open, content, close, stretchy)
	n.TeXClass = mml.TeXClassInner
	return n
}

// AutoOpen fences use the node factory; changing to p.token would alter the
// accepted construction. Content has already supplied its own creation events.
func (p *parser) autoOpenFence(text string, class mml.TeXClass) *mml.Node {
	return p.noteMO(autoOpenFence(text, class))
}

func (p *parser) normalizeDecorationBase(base *mml.Node) *mml.Node {
	return normalizeDecorationBaseWithMO(base, func() *mml.Node { return p.noteMO(node("mo")) })
}
