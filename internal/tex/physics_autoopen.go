// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on PhysicsMethods.Derivative, PhysicsItems.AutoOpen and ParseUtil.fenced.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// commandResult carries one invocation's stack item and after-node action.
// Neither lives in shared parseState or escapes a genuine child parser.
// Recipients deliver nodes before activating afterNode.
type commandResult struct {
	nodes         []*mml.Node
	namedFunction bool
	notItem       bool
	dotsItem      *pendingDots
	afterNode     *derivativeAutoOpen
}

func (p *parser) commandEvent(name string) (result commandResult, err error) {
	namedFunction, notItem := p.commandNamedFunction, p.commandNot
	dotsItem := p.commandDots
	p.commandNamedFunction, p.commandNot = false, false
	p.commandDots = nil
	defer func() {
		p.commandNamedFunction, p.commandNot = namedFunction, notItem
		p.commandDots = dotsItem
	}()
	result.nodes, err = p.commandNodes(name, &result.afterNode)
	result.namedFunction, result.notItem = p.commandNamedFunction, p.commandNot
	result.dotsItem = p.commandDots
	return result, err
}

// Direct command callers have no pending recipient. The row and script
// consumers use commandEvent so they can deliver the initial nodes first.
func (p *parser) command(name string) ([]*mml.Node, error) {
	result, err := p.commandEvent(name)
	if err != nil {
		return nil, err
	}
	if result.notItem {
		result.nodes = append(result.nodes, notFallback())
	}
	if result.dotsItem != nil {
		result.nodes = append(result.nodes, result.dotsItem.finish()...)
	}
	tail, err := result.afterNode.complete(p)
	return append(result.nodes, tail...), err
}

type derivativeAutoOpen struct {
	ignore    bool
	openCount int
	closed    bool
}

func (a *derivativeAutoOpen) start(p *parser) bool {
	if a == nil {
		return false
	}
	// TexParser.GetNext uses ECMAScript whitespace; this is not a change to
	// ordinary character scanning or argument readers.
	for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
		p.consumeRune()
	}
	if p.pos == len(p.source) || p.source[p.pos] != '(' {
		return false
	}
	p.pos++
	return true
}

func (a *derivativeAutoOpen) complete(p *parser) ([]*mml.Node, error) {
	return a.completeAfter(p, nil)
}

// A real AutoOpen is a non-MML successor. Notify its recipient only if it
// starts, after the command's initial nodes but before parsing its body.
func (a *derivativeAutoOpen) completeAfter(p *parser, before func()) ([]*mml.Node, error) {
	if !a.start(p) {
		return nil, nil
	}
	if before != nil {
		before()
	}
	content, _, err := p.parseRowWithAutoOpen(0, false, false, a)
	if err != nil {
		return nil, err
	}
	if a.ignore {
		return nil, nil
	}
	// AutoOpen.toMml delegates to fenced, then removes the row's open/close/
	// texClass properties. Fence nodes use the node factory, not token factory.
	children := []*mml.Node{p.autoOpenFence("(", mml.TeXClassOpen)}
	children = append(children, content...)
	children = append(children, p.autoOpenFence(")", mml.TeXClassClose))
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
