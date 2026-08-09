// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func (w *wrapper) actionOffsetX() float64 {
	offsets := strings.Fields(stringAttribute(w.node, "data-offsets", ""))
	dx := ".2em"
	if len(offsets) != 0 {
		dx = offsets[0]
	}
	return layout.Length2Em(dx, .2, w.bbox.Scale, w.renderer.pxPerEm)
}

func (w *wrapper) actionSelection() int {
	selection := 1.0
	if value, ok := numberAttribute(w.node, "selection"); ok {
		selection = value
	}
	selection = math.Max(1, math.Min(float64(len(w.children)), selection))
	return int(selection) - 1
}

func (w *wrapper) selectedActionChild() *wrapper {
	index := w.actionSelection()
	if index >= 0 && index < len(w.children) {
		return w.children[index]
	}
	empty := mml.NewNode("mrow", nil, nil)
	return w.renderer.wrap(empty, w, w.scriptLevel, w.displayStyle)
}

func (w *wrapper) computeActionBBox(bbox *layout.BBox) {
	w.dx = w.actionOffsetX()
	bbox.UpdateFrom(w.selectedActionChild().outerBBox())
}

func (w *wrapper) actionSelectionText() string {
	selection := 1.0
	if value, ok := numberAttribute(w.node, "selection"); ok {
		selection = value
	}
	return strconv.FormatFloat(selection, 'f', -1, 64)
}

func (w *wrapper) actionToSVG(parent *Element) {
	w.dx = w.actionOffsetX()
	element := w.standardSVG(parent)
	selected := w.selectedActionChild()
	bbox := selected.outerBBox()
	rect := NewElement("rect").
		SetAttr("width", fixed(bbox.W)).
		SetAttr("height", fixed(bbox.H+bbox.D)).
		SetAttr("y", fixed(-bbox.D)).
		SetAttr("fill", "none").
		SetAttr("pointer-events", "all")
	element.Append(rect)
	selected.toSVG(element)
	selected.place(bbox.L*bbox.RScale, 0, selected.element)

	action := stringAttribute(w.node, "actiontype", "toggle")
	switch action {
	case "toggle":
		// LiteElement has no event API.  Keep the exact static marker; click
		// selection and rerendering are intentionally unavailable in D2.
		element.SetAttr("data-toggle", w.actionSelectionText())
	case "tooltip":
		if len(w.children) > 1 && w.children[1].node.Kind == "mtext" {
			element.Prepend(NewElement("title", Text(nodeText(w.children[1].node))))
		}
	case "statusline":
		if len(w.children) > 1 && w.children[1].node.Kind == "mtext" {
			element.SetAttr("data-statusline", nodeText(w.children[1].node))
		}
	}
}
