// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func actionFixture(action string, selection any, children ...*mml.Node) *mml.Node {
	node := mml.NewNode("maction", nil, nil, children...)
	node.TeXClass = mml.TeXClassOrd
	node.Attributes.Set("actiontype", action)
	if selection != nil {
		node.Attributes.Set("selection", selection)
	}
	return node
}

func TestActionTextTooltipFrozenSourceShapedSVG(t *testing.T) {
	selected := wrapperTrancheToken("mi", "x", mml.TeXClassOrd)
	tip := wrapperTrancheToken("mtext", "tip", mml.TeXClassOrd)
	node := actionFixture("tooltip", nil, selected, tip)
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(node))
	const want = "cfd08d86a945d97436cf2c3db4d1869e73ece179e39ac8c2f2c89dc600a90e76"
	requireFrozenWrapperHash(t, got, want)
}

func TestActionSelectionToggleAndStatusMarkers(t *testing.T) {
	first := wrapperTrancheToken("mi", "x", mml.TeXClassOrd)
	second := wrapperTrancheToken("mi", "y", mml.TeXClassOrd)
	node := actionFixture("toggle", 9, first, second)
	node.Attributes.Set("data-offsets", ".3em .4em")
	w := wrapperTrancheWrapper(node)
	w.computeActionBBox(w.bbox)
	w.bboxComputed = true
	root := NewElement("root")
	w.actionToSVG(root)
	output := root.String()
	for _, fragment := range []string{
		`data-toggle="9"`,
		`data-c="1D466"`, // clamped selection renders y
		`pointer-events="all"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("toggle output is missing %q:\n%s", fragment, output)
		}
	}
	if w.dx != .3 {
		t.Errorf("data-offsets dx = %g, want .3", w.dx)
	}

	status := actionFixture("statusline", nil, first.Clone(), wrapperTrancheToken("mtext", "ready", mml.TeXClassOrd))
	w = wrapperTrancheWrapper(status)
	w.computeActionBBox(w.bbox)
	w.bboxComputed = true
	root = NewElement("root")
	w.actionToSVG(root)
	if !strings.Contains(root.String(), `data-statusline="ready"`) {
		t.Fatalf("statusline marker is missing:\n%s", root.String())
	}
}

func TestActionLiteDOMUnsupportedMathTooltipIsStatic(t *testing.T) {
	node := actionFixture("tooltip", nil,
		wrapperTrancheToken("mi", "x", mml.TeXClassOrd),
		mml.NewNode("mfrac", nil, nil,
			wrapperTrancheToken("mn", "1", mml.TeXClassOrd),
			wrapperTrancheToken("mn", "2", mml.TeXClassOrd),
		),
	)
	w := wrapperTrancheWrapper(node)
	w.computeActionBBox(w.bbox)
	w.bboxComputed = true
	root := NewElement("root")
	w.actionToSVG(root)
	output := root.String()
	if strings.Contains(output, "<title>") || strings.Contains(output, "foreignObject") {
		t.Fatalf("browser-only math tooltip leaked into Lite DOM output:\n%s", output)
	}
}
