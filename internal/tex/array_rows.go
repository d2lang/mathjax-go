// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/base/BaseItems.ts ArrayItem.EndRow/EndTable.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// An explicit row separator always closes the pending entry, even if it is
// empty. EndTable closes it only when nodes or previously closed entries
// remain. Each mtd owns an inferred content row: an empty explicit group,
// style, or leaf mspace is still an emitted node and must not be discarded.
func omitFinalArrayRow(index, rowCount int, cells []*mml.Node) bool {
	return index == rowCount-1 && emptyArrayCells(cells)
}

func emptyArrayCells(cells []*mml.Node) bool {
	if len(cells) != 1 {
		return len(cells) == 0
	}
	for _, content := range cells[0].Children {
		if content.Kind != "mrow" || !content.Flags.Inferred || len(content.Children) != 0 {
			return false
		}
	}
	return true
}
