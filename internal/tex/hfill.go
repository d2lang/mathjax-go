// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/base/{BaseMethods,BaseItems}.ts.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// HFill records ArrayItem.Size without emitting an MML node. SetFont changes
// the same environment, so its continuation retains the preceding node count.
type arrayCellState struct {
	fills    []int
	offset   int
	equation *equationRowState
	multline bool
	numCases bool
	shove    string
	rules    *arrayRules
}

func isHFill(name string) bool {
	return name == "hfill" || name == "hfil" || name == "hfilll"
}

func unsupportedHFill(name string) error {
	return texError("UnsupportedHFill", "Unsupported use of %s", "\\"+name)
}

const arrayCellAlignment = "arrayCellAlignment"

func (s *arrayCellState) finish(content *mml.Node, size int) *mml.Node {
	if s == nil || len(s.fills) == 0 {
		return content
	}
	align := ""
	if s.fills[0] == 0 {
		align = "right"
	}
	if s.fills[len(s.fills)-1] == size {
		if align != "" {
			align = "center"
		} else {
			align = "left"
		}
	}
	if align != "" {
		content.SetProperty(arrayCellAlignment, align)
	}
	return content
}

func arrayCellNode(content *mml.Node) *mml.Node {
	cell := node("mtd", content)
	if alignment, ok := content.Property(arrayCellAlignment); ok {
		cell.Attributes.Set("columnalign", alignment)
	}
	return cell
}

func (p *parser) parseArrayCellString(source string, rules ...*arrayRules) (*mml.Node, error) {
	cell := &arrayCellState{}
	if len(rules) != 0 {
		cell.rules = rules[0]
	}
	return p.parseStringWithEnvironment(source, p.ensureStackGlobal(), nil, cell, p.environmentOwner)
}
