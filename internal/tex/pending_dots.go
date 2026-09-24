// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Dots, BaseItems.DotsItem and MmlMo.texClass (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// DotsItem keeps both candidates as properties, not stack children. Create
// both now, including the unused one, in the original token-factory order.
type pendingDots struct {
	ldots *mml.Node
	cdots *mml.Node
}

func (p *parser) startDots() *pendingDots {
	ldots := p.token("mo", "\u2026")
	ldots.Attributes.Set("stretchy", false)
	p.applyVectorFactory(ldots)
	cdots := p.token("mo", "\u22EF")
	cdots.Attributes.Set("stretchy", false)
	p.applyVectorFactory(cdots)
	return &pendingDots{ldots: ldots, cdots: cdots}
}

func (pending *pendingDots) active() bool { return pending.ldots != nil }

// A non-MML successor (including a row end) always selects lower dots.
func (pending *pendingDots) finish() []*mml.Node {
	if !pending.active() {
		return nil
	}
	dots := pending.ldots
	*pending = pendingDots{}
	return []*mml.Node{dots}
}

// Empty PushAll results leave DotsItem pending. An actual MML successor selects
// the candidate, then is delivered unchanged after it. Groups and left/right
// expressions arrive here only after their enclosing item has completed.
func (pending *pendingDots) apply(nodes []*mml.Node) []*mml.Node {
	if !pending.active() || len(nodes) == 0 {
		return nodes
	}
	dots := pending.ldots
	if core := dotsCoreMO(nodes[0]); core != nil {
		class := dotsOperatorClass(core)
		if class == mml.TeXClassBin || class == mml.TeXClassRel {
			dots = pending.cdots
		}
	}
	*pending = pendingDots{}
	return append([]*mml.Node{dots}, nodes...)
}

func dotsCoreMO(n *mml.Node) *mml.Node {
	if n == nil || !n.Flags.Embellished {
		return nil
	}
	for n.Kind != "mo" {
		next := n.Core()
		if next == nil || next == n {
			return nil
		}
		n = next
	}
	return n
}

// Before inheritance, a mo's cached Go class can be only its constructor's
// provisional value. The primary getter instead honors an explicitly assigned
// class or looks up the current forms, defaulting to REL without saving it.
// Do not run operator inheritance here: it also changes attributes and spacing.
func dotsOperatorClass(n *mml.Node) mml.TeXClass {
	if _, explicit := n.Property("texClass"); explicit {
		return n.TeXClass
	}
	forms := operatorForms(n)
	if n.Attributes.IsSet("form") {
		value, _ := n.Attributes.Get("form")
		if form, ok := value.(string); ok {
			ordered := []string{form}
			for _, candidate := range forms {
				if candidate != form {
					ordered = append(ordered, candidate)
				}
			}
			forms = ordered
		}
	}
	if definition, ok := lookupOperatorDefinition(textContent(n), forms); ok {
		return mml.TeXClass(definition.TexClass)
	}
	return mml.TeXClassRel
}
