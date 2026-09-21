// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/output/common/Wrappers/ms.ts and ts/output/svg/Wrappers/ms.ts.

package svg

import (
	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/mml"
)

// CommonMs adds quote text to its wrappers, never to the original MathML.
// An explicitly supplied or directly inherited quote remains literal. Only
// unset default quotes become curly, and monospace always keeps them straight.
func (w *wrapper) initializeStringQuotes() {
	left := stringAttribute(w.node, "lquote", `"`)
	right := stringAttribute(w.node, "rquote", `"`)
	if w.variant != font.Monospace {
		if !w.node.Attributes.IsSet("lquote") && left == `"` {
			left = "\u201C"
		}
		if !w.node.Attributes.IsSet("rquote") && right == `"` {
			right = "\u201D"
		}
	}
	open := w.renderer.wrap(mml.NewText(left), w, w.scriptLevel, w.displayStyle)
	close := w.renderer.wrap(mml.NewText(right), w, w.scriptLevel, w.displayStyle)
	w.children = append([]*wrapper{open}, w.children...)
	w.children = append(w.children, close)
}
