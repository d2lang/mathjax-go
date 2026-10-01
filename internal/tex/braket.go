// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/braket/{BraketMethods,BraketItems}.ts.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

type braketItem struct {
	stretchy, single bool
	barcount, barmax int
	open, close      string
	singleFence      *mml.Node
}

// Braket parses on the caller: an unbraced control sequence may consume its
// own arguments or install a macro expansion before delivering its first node.
func (p *parser) braket(name string) ([]*mml.Node, error) {
	p.skipNextSpaces()
	if p.pos == len(p.source) {
		return nil, texError("MissingArgFor", "Missing argument for %s", "\\"+name)
	}
	item := &braketItem{stretchy: name != "set", single: true, barmax: -1}
	open, close := "⟨", "⟩"
	if name == "Set" || name == "set" {
		open, close = "{", "}"
		item.barmax = 1
	}
	item.open, item.close = open, close
	terminator := byte(0)
	if p.source[p.pos] == '{' {
		p.pos++
		item.single = false
		terminator = '}'
	}
	previous, cdArrayEntry, environmentRow, ordinaryArray := p.braketOwner, p.cdArrayEntry, p.environmentRow, p.ordinaryArray
	p.environmentRow, p.ordinaryArray = nil, nil
	p.braketOwner = item
	if !item.single {
		p.cdArrayEntry = false
	}
	defer func() {
		p.braketOwner, p.cdArrayEntry, p.environmentRow, p.ordinaryArray = previous, cdArrayEntry, environmentRow, ordinaryArray
	}()
	children, _, err := p.parseRowWithInfix(terminator, false, false)
	if err != nil {
		return nil, err
	}
	popped := p.environmentPopped
	if popped {
		// SpreadLines calls the popped BraketItem's toMml(), which still
		// constructs its fences even without a matching close item.
		p.environmentPopped = false
	}
	if item.single && p.pendingCell != nil {
		// A final MML closed this Braket before the CellItem was replayed.
		// Its copied lexical environment is now gone; the enclosing open
		// owner must save its own environment when it receives the cell.
		p.pendingCell.envSaved = false
	}
	var tail []*mml.Node
	if item.single && len(children) > 1 {
		tail = children[1:]
		children = children[:1]
	}
	// NodeUtil.appendChildren flattens the inferred inner row when the
	// BraketItem constructs its explicit fenced row.
	result := item.singleFence
	if result == nil {
		result = p.leftRightFenced(open, row(children, true), close, item.stretchy)
	}
	if popped {
		if err := mathtoolsSpreadPop(result, nil); err != nil {
			return nil, err
		}
	}
	return append([]*mml.Node{result}, tail...), nil
}

func (p *parser) braketBar() []*mml.Node {
	owner := p.braketOwner
	if owner == nil || owner.barmax >= 0 && owner.barcount >= owner.barmax {
		return []*mml.Node{p.operator("∥", mml.TeXClassOrd, map[string]any{"stretchy": false})}
	}
	if !owner.stretchy {
		// Fixed sets do not increment barcount.
		return []*mml.Node{setAttributes(p.token("mo", "∥"), map[string]any{"stretchy": false, "braketbar": true})}
	}
	closing := texAtom(row(nil, true), mml.TeXClassClose)
	// Delivering CLOSE completes a single owner before the bar is created.
	// Preserve that order in the parse's operator-creation list as well as
	// keeping the remaining bar and OPEN outside the resulting fences.
	if owner.single {
		owner.singleFence = p.leftRightFenced(owner.open, closing, owner.close, true)
		p.braketOwner = nil
	}
	owner.barcount++
	bar := setAttributes(p.token("mo", "∥"), map[string]any{"stretchy": true, "braketbar": true})
	return []*mml.Node{closing, bar, texAtom(row(nil, true), mml.TeXClassOpen)}
}
