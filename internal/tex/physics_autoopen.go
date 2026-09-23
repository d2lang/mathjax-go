// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on PhysicsMethods.Derivative, PhysicsItems.AutoOpen and ParseUtil.fenced.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// commandResult owns a single invocation's after-node action. No action lives
// in shared parseState or escapes a genuine child parser. Recipients deliver
// nodes (including a pending script attachment) before activating afterNode.
type commandResult struct {
	nodes         []*mml.Node
	namedFunction bool
	afterNode     *derivativeAutoOpen
}

func (p *parser) commandEvent(name string) (result commandResult, err error) {
	p.commandNamedFunction = false
	result.nodes, err = p.commandNodes(name, &result.afterNode)
	result.namedFunction = p.commandNamedFunction
	p.commandNamedFunction = false
	return result, err
}

// Direct command callers have no pending recipient. The row and script
// consumers use commandEvent so they can deliver the initial nodes first.
func (p *parser) command(name string) ([]*mml.Node, error) {
	result, err := p.commandEvent(name)
	if err != nil {
		return nil, err
	}
	tail, err := result.afterNode.complete(p)
	return append(result.nodes, tail...), err
}

type derivativeAutoOpen struct {
	ignore    bool
	openCount int
	closed    bool
}

func (a *derivativeAutoOpen) complete(p *parser) ([]*mml.Node, error) {
	if a == nil {
		return nil, nil
	}
	// TexParser.GetNext uses ECMAScript whitespace; this is not a change to
	// ordinary character scanning or argument readers.
	for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
		p.consumeRune()
	}
	if p.pos == len(p.source) || p.source[p.pos] != '(' {
		return nil, nil
	}
	p.pos++
	content, _, err := p.parseRowWithAutoOpen(0, false, false, a)
	if err != nil {
		return nil, err
	}
	if a.ignore {
		return nil, nil
	}
	// AutoOpen.toMml delegates to fenced, then removes the row's open/close/
	// texClass properties. Fence nodes use the node factory, not token factory.
	children := []*mml.Node{autoOpenFence("(", mml.TeXClassOpen)}
	children = append(children, content...)
	children = append(children, autoOpenFence(")", mml.TeXClassClose))
	return []*mml.Node{forcedRow(children, false)}, nil
}

func autoOpenFence(text string, class mml.TeXClass) *mml.Node {
	n := node("mo", mml.NewText(text))
	n.Attributes.Set("fence", true)
	n.Attributes.Set("stretchy", true)
	n.Attributes.Set("symmetric", true)
	n.TeXClass = class
	n.SetProperty("texClass", class)
	return n
}

func (a *derivativeAutoOpen) observe(nodes []*mml.Node) {
	if len(nodes) == 1 && nodes[0].Kind == "mo" && textContent(nodes[0]) == "(" {
		a.openCount++
	}
}

func (a *derivativeAutoOpen) close() bool {
	before := a.openCount
	a.openCount-- // Preserve the source's zero-to-minus-one closing transition.
	if before == 0 {
		a.closed = true
	}
	return a.closed
}

func (a *derivativeAutoOpen) stopError() error {
	return texError("ExtraOrMissingDelims", "Extra open or missing close delimiter")
}
