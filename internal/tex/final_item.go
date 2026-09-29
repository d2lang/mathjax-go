// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on Stack.Push, PhysicsMethods.AutoClose and BaseItems.PositionItem.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// autoclose belongs to the source MmlItem, not its MML node. A recipient may
// replay the same item after changing its node, or replace it with a new item.
type finalItem struct {
	node      *mml.Node
	autoclose byte
}

func finalItems(nodes []*mml.Node) []finalItem {
	items := make([]finalItem, 0, len(nodes))
	for _, n := range nodes {
		if n != nil {
			items = append(items, finalItem{node: n})
		}
	}
	return items
}

func (item *positionItem) applyFinal(incoming finalItem) []finalItem {
	if item.left != nil {
		// Horizontal Position replays the identical incoming MmlItem.
		return []finalItem{{node: item.left}, incoming, {node: item.right}}
	}
	// Vertical Position creates a new item around mpadded; no item properties
	// are copied from the consumed incoming item.
	return finalItems(item.apply(incoming.node))
}
