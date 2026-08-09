// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/amscd/AmsCdMethods.ts and AmsCdConfiguration.ts.

package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
)

const (
	amscdMinWidthState  = "\x00amscd-min-width"
	amscdMinHeightState = "\x00amscd-min-height"
)

// amscdCommand handles the two source commands that set per-parser CD arrow
// dimensions.  Values live in parseState so they persist until compilation
// ends, matching MathJax's stack environment without adding shared fields.
func (p *parser) amscdCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	key := ""
	switch name {
	case "minCDarrowwidth":
		key = amscdMinWidthState
	case "minCDarrowheight":
		key = amscdMinHeightState
	default:
		return nil, false, nil
	}
	value, err := p.readDimension(name)
	if err != nil {
		return nil, true, err
	}
	p.state.macros[key] = macroDefinition{body: value}
	return nil, true, nil
}

// amscdEnvironment wraps the already source-shaped CD parser to install the
// exact package table options and arrow minsize attributes.
func (p *parser) amscdEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if name != "CD" {
		return nil, false, nil
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}
	nodes, err = p.parseCD(body)
	if err != nil {
		return nil, true, err
	}
	if len(nodes) == 0 {
		return nodes, true, nil
	}
	table := nodes[0]
	for _, row := range table.Children {
		if len(row.Children) != 0 && amscdCellHasVerticalArrow(row.Children[len(row.Children)-1]) {
			row.AppendChild(node("mtd"))
		}
	}
	resetTableAttributes(table,
		"columnalign", "center",
		"columnspacing", "5pt",
		"rowspacing", "5pt",
		"displaystyle", true,
	)
	minWidth := p.amscdState(amscdMinWidthState, "2.75em")
	minHeight := p.amscdState(amscdMinHeightState, "1.75em")
	table.Walk(func(current *mml.Node) bool {
		if current.Kind != "mo" {
			return true
		}
		text := textContent(current)
		switch text {
		case "→", "←", "=":
			amscdResetAttributes(current,
				"minsize", minWidth,
				"stretchy", true,
			)
		case "↓", "↑", "‖":
			amscdResetAttributes(current,
				"minsize", minHeight,
				"stretchy", true,
				"symmetric", true,
				"lspace", 0,
				"rspace", 0,
			)
		}
		return true
	})
	return nodes, true, nil
}

func amscdCellHasVerticalArrow(cell *mml.Node) bool {
	found := false
	cell.Walk(func(current *mml.Node) bool {
		if current.Kind == "mo" {
			switch textContent(current) {
			case "↓", "↑", "‖":
				found = true
			}
		}
		return !found
	})
	return found
}

func (p *parser) amscdState(key, fallback string) string {
	if definition, ok := p.state.macros[key]; ok {
		return definition.body
	}
	return fallback
}

func amscdResetAttributes(node *mml.Node, entries ...any) {
	for _, name := range node.Attributes.ExplicitNames() {
		node.Attributes.Explicit().Delete(name)
	}
	for i := 0; i < len(entries); i += 2 {
		node.Attributes.Set(entries[i].(string), entries[i+1])
	}
}
