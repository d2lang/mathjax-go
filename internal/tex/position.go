// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.RaiseLower/MoveLeftRight and BaseItems.PositionItem.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

type positionItem struct {
	name, height, depth string
	left, right         *mml.Node
}

func (item *positionItem) apply(child *mml.Node) []*mml.Node {
	if item.left != nil {
		// PositionItem returns three separate final items. An enclosing
		// position consumes only the first spacer, not this entire sequence.
		return []*mml.Node{item.left, child, item.right}
	}
	padded := node("mpadded", child)
	padded.Attributes.Set("height", item.height)
	padded.Attributes.Set("depth", item.depth)
	padded.Attributes.Set("voffset", item.height)
	return []*mml.Node{padded}
}

func (item *positionItem) missingBox() error {
	return texError("MissingBoxFor", "Missing box for \\%s", item.name)
}

type rowPositionFrame struct {
	item       *positionItem
	prefix     []*mml.Node
	styleDepth int
}

// CrLaTeX reads its optional dimension before delivering the close item that
// can make PositionItem report a missing box. Preserve that error precedence.
func (p *parser) positionLinebreak(name string) error {
	if name != "\\" {
		return nil
	}
	if p.pos < len(p.source) && p.source[p.pos] == '*' {
		p.pos++
	}
	if p.pos >= len(p.source) || p.source[p.pos] != '[' {
		return nil
	}
	dimension, _, err := p.readBrackets(nil)
	if err != nil {
		return err
	}
	if dimension != "" && !dimensionFull.MatchString(dimension) {
		return texError("BracketMustBeDimension", "Bracket argument to \\%s must be a dimension", name)
	}
	return nil
}
