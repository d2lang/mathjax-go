// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: BaseMethods.Entry/Cr/CrLaTeX, BaseItems.CellItem, and StackItem.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// A cell is a closing stack item, not a space or a pre-parsed row boundary.
// It can be replayed after Over/Style or a completed single Braket before it
// reaches the actual open owner. CrLaTeX's spaces are delivered afterward.
type cellItem struct {
	name      string
	cr        bool
	linebreak bool
	spacing   string
	font      string
	emptyFont bool
	color     string
	envSaved  bool
	// Cases.Entry observes this before pushing the closing CellItem.
	numCasesText bool
}

func (p *parser) crCommand(name string) error {
	item := &cellItem{name: "\\" + name, cr: true, linebreak: name != "cr"}
	if name == "\\" {
		// CrLaTeX tests raw characters here, without GetNext's whitespace.
		if p.pos < len(p.source) && p.source[p.pos] == '*' {
			p.pos++
		}
		if p.pos < len(p.source) && p.source[p.pos] == '[' {
			dimension, _, err := p.readBrackets(nil)
			if err != nil {
				if e, ok := err.(*Error); ok && e.ID == "MissingCloseBracket" {
					return texError(e.ID, "Could not find closing ']' for argument to %s", item.name)
				}
				return err
			}
			if dimension != "" {
				match := dimensionFull.FindStringSubmatch(dimension)
				if match == nil {
					return texError("BracketMustBeDimension", "Bracket argument to %s must be a dimension", item.name)
				}
				item.spacing = dimensionValue(match)
			}
		}
	}
	p.commandCell = item
	return nil
}

func (item *cellItem) misplaced() error {
	return texError("Misplaced", "Misplaced %s", item.name)
}

func (item *cellItem) saveEnvironment(p *parser) {
	if !item.envSaved {
		item.font, item.emptyFont, item.color = p.activeFont, p.fontExplicitEmpty, p.activeColor
		item.envSaved = true
	}
}

func (item *cellItem) spaces() []*mml.Node {
	var nodes []*mml.Node
	if item.spacing != "" {
		nodes = append(nodes, setAttributes(node("mspace"), map[string]any{"depth": item.spacing}))
	}
	return append(nodes, setAttributes(node("mspace"), map[string]any{"linebreak": "newline"}))
}
