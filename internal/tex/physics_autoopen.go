// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on PhysicsMethods.Derivative/OperatorApplication, PhysicsItems.AutoOpen
// and ParseUtil.fenced.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// commandResult carries one invocation's stack item and after-node action.
// Neither lives in shared parseState or escapes a genuine child parser.
// Recipients deliver nodes before activating afterNode.
type commandResult struct {
	nodes         []*mml.Node
	namedFunction bool
	notItem       bool
	nonscriptItem bool
	dotsItem      *pendingDots
	positionItem  *positionItem
	cellItem      *cellItem
	afterNode     *derivativeAutoOpen
}

func (p *parser) commandEvent(name string) (result commandResult, err error) {
	namedFunction, notItem := p.commandNamedFunction, p.commandNot
	nonscript := p.commandNonscript
	dotsItem, position := p.commandDots, p.commandPosition
	cell := p.commandCell
	p.commandNamedFunction, p.commandNot = false, false
	p.commandNonscript = false
	p.commandDots, p.commandPosition = nil, nil
	p.commandCell = nil
	defer func() {
		p.commandNamedFunction, p.commandNot = namedFunction, notItem
		p.commandNonscript = nonscript
		p.commandDots, p.commandPosition = dotsItem, position
		p.commandCell = cell
	}()
	result.nodes, err = p.commandNodes(name, &result.afterNode)
	result.namedFunction, result.notItem = p.commandNamedFunction, p.commandNot
	result.nonscriptItem = p.commandNonscript
	result.dotsItem, result.positionItem = p.commandDots, p.commandPosition
	result.cellItem = p.commandCell
	return result, err
}

// Direct command callers have no pending recipient. The row and script
// consumers use commandEvent so they can deliver the initial nodes first.
func (p *parser) command(name string) ([]*mml.Node, error) {
	result, err := p.commandEvent(name)
	if err != nil {
		return nil, err
	}
	if result.cellItem != nil {
		if !result.cellItem.linebreak {
			return nil, result.cellItem.misplaced()
		}
		result.nodes = append(result.nodes, result.cellItem.spaces()...)
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
	open        byte
	closer      byte
	application *physicsApplicationArgument
	ignore      bool
	openCount   int
	closed      bool
}

func (a *derivativeAutoOpen) openingFence() byte {
	if a.open == 0 {
		return '('
	}
	return a.open
}

func (a *derivativeAutoOpen) closingFence() byte {
	if a.closer == 0 {
		return ')'
	}
	return a.closer
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
	if p.pos == len(p.source) || p.source[p.pos] != a.openingFence() {
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
	if a != nil && a.application != nil {
		opened, err := a.application.prepare(p, a)
		if err != nil || !opened {
			return nil, err
		}
	}
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
	children := []*mml.Node{p.autoOpenFence(string(a.openingFence()), mml.TeXClassOpen)}
	children = append(children, content...)
	children = append(children, p.autoOpenFence(string(a.closingFence()), mml.TeXClassClose))
	// Removing the texClass property retains the class assigned by fenced.
	result := forcedRow(children, false)
	result.TeXClass = mml.TeXClassInner
	return []*mml.Node{result}, nil
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
	if len(nodes) == 1 && nodes[0].Kind == "mo" && textContent(nodes[0]) == string(a.openingFence()) {
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
