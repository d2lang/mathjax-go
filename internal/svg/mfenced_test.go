// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func fencedFixture() *mml.Node {
	node := mml.NewNode("mfenced", nil, nil,
		wrapperTrancheToken("mi", "x", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "y", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "z", mml.TeXClassOrd),
	)
	node.Attributes.Set("open", " [ ")
	node.Attributes.Set("close", " ] ")
	node.Attributes.Set("separators", "; |")
	node.TeXClass = mml.TeXClassInner
	return node
}

func TestFencedFrozenSourceShapedSVG(t *testing.T) {
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(fencedFixture()))
	const want = "8766038a58b96b9d808776a05989b3e0f3d92d2c50a32245fbe041ea71700fd7"
	requireFrozenWrapperHash(t, got, want)
}

func TestFencedDefaultsWhitespaceAndRepeatedSeparator(t *testing.T) {
	node := mml.NewNode("mfenced", nil, nil,
		wrapperTrancheToken("mi", "a", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "b", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "c", mml.TeXClassOrd),
	)
	node.Attributes.Set("open", " \t(\n")
	node.Attributes.Set("close", "\r) ")
	node.Attributes.Set("separators", " ; ")
	w := wrapperTrancheWrapper(node)
	w.computeFencedBBox(w.bbox)
	w.bboxComputed = true
	root := NewElement("root")
	w.fencedToSVG(root)
	output := root.String()
	if strings.Count(output, `data-c="3B"`) != 2 {
		t.Fatalf("last separator was not repeated:\n%s", output)
	}
	for _, code := range []string{`data-c="28"`, `data-c="29"`} {
		if !strings.Contains(output, code) {
			t.Fatalf("whitespace-stripped fence %s is missing:\n%s", code, output)
		}
	}
}
