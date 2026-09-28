// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/ParseUtil.ts fixInitialMO; base/BaseItems.ts
// EqnArrayItem.EndEntry and ams/AmsItems.ts MultlineItem.EndEntry.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// fixInitialMO preserves operator spacing at an alignment boundary. Source
// skips mspace and empty TeXAtoms, then inspects the first remaining node.
// Every embellished operator qualifies, as does a relation TeXAtom.
func fixInitialMO(children []*mml.Node) []*mml.Node {
	for _, child := range children {
		if child == nil || child.Kind == "mspace" {
			continue
		}
		if child.Kind == "TeXAtom" && (len(child.Children) == 0 || len(child.Children[0].Children) == 0) {
			continue
		}
		if child.Flags.Embellished || (child.Kind == "TeXAtom" && child.TeXClass == mml.TeXClassRel) {
			return append([]*mml.Node{node("mi")}, children...)
		}
		break
	}
	return children
}

// EqnArray applies the repair to every entry after the first in each row.
// Multline applies it to each entry after the first row. Call before adding
// label cells or layout padding that was not present at EndEntry.
func fixInitialArrayOperators(table *mml.Node, firstRow, firstColumn int) {
	if len(table.Children) <= firstRow {
		return
	}
	for _, row := range table.Children[firstRow:] {
		if len(row.Children) <= firstColumn {
			continue
		}
		for _, cell := range row.Children[firstColumn:] {
			if len(cell.Children) == 0 {
				continue
			}
			content := cell.Children[0]
			content.SetChildren(fixInitialMO(content.Children))
		}
	}
}
