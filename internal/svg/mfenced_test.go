// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func fencedFixture() *mml.Node {
	node := syntheticMMLFactory.Create("mfenced",
		wrapperTrancheToken("mi", "x", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "y", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "z", mml.TeXClassOrd),
	)
	node.Attributes.Set("open", " [ ")
	node.Attributes.Set("close", " ] ")
	node.Attributes.Set("separators", "; |")
	node.TeXClass = mml.TeXClassInner
	initializeFrozenFencedTestNodes(node)
	return node
}

// These low-level wrapper fixtures have no TeX inheritance stage. Initialize
// their limited fixed-character fake nodes from the original inherited states;
// the complete Go inheritance path is checked by the constructed-MathML tests
// in internal/tex, including font, size and incoming/explicit-level conflicts.
func initializeFrozenFencedTestNodes(node *mml.Node) {
	syntheticMMLFactory.CreateFencedNodes(node)
	for _, fake := range append([]*mml.Node{node.Fenced.Open, node.Fenced.Close}, node.Fenced.Separators...) {
		if fake == nil {
			continue
		}
		fake.Attributes.SetInherited("displaystyle", false)
		fake.Attributes.SetInherited("scriptlevel", 0)
		form := stringAttribute(fake, "form", "infix")
		fake.Attributes.SetInherited("form", form)
		switch nodeText(fake) {
		case "[", "(", "]", ")", "|":
			fake.Attributes.SetInherited("fence", true)
			fake.Attributes.SetInherited("stretchy", true)
			fake.Attributes.SetInherited("symmetric", true)
			if form == "infix" {
				fake.TeXClass = mml.TeXClassOrd
			}
		case ";":
			fake.Attributes.SetInherited("linebreakstyle", "after")
			fake.Attributes.SetInherited("separator", true)
			fake.TeXClass = mml.TeXClassPunct
		}
	}
}

func TestFencedFrozenSourceShapedSVG(t *testing.T) {
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(fencedFixture()))
	const want = "8766038a58b96b9d808776a05989b3e0f3d92d2c50a32245fbe041ea71700fd7"
	requireFrozenWrapperHash(t, got, want)
}

func TestFencedDefaultsWhitespaceAndRepeatedSeparator(t *testing.T) {
	node := syntheticMMLFactory.Create("mfenced",
		wrapperTrancheToken("mi", "a", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "b", mml.TeXClassOrd),
		wrapperTrancheToken("mi", "c", mml.TeXClassOrd),
	)
	node.Attributes.Set("open", " \t(\n")
	node.Attributes.Set("close", "\r) ")
	node.Attributes.Set("separators", " ; ")
	initializeFrozenFencedTestNodes(node)
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

func TestFencedWrapperRowOwnershipAndRepeat(t *testing.T) {
	node := fencedFixture()
	w := wrapperTrancheWrapper(node)
	row := w.fencedRow
	if row == nil || len(row.node.Children) != 0 || row.node.Parent != nil {
		t.Fatal("temporary wrapper row acquired semantic children or a parent")
	}
	if !row.node.Flags.Spacelike || row.node.Flags.Embellished || row.node.Core() != row.node {
		t.Fatal("empty temporary row lost its registered dynamic state")
	}
	first := *w.getBBox()
	for i := 0; i < 2; i++ {
		w.bboxComputed = false
		if got := *w.getBBox(); got != first || w.fencedRow != row {
			t.Fatal("repeated layout changed the cached fenced row")
		}
		w.toSVG(NewElement("g"))
		for j, child := range node.Children {
			if child.Parent != node || w.children[j].parent != w {
				t.Fatal("authored child ownership was not restored")
			}
		}
		for _, fake := range append([]*mml.Node{node.Fenced.Open, node.Fenced.Close}, node.Fenced.Separators...) {
			if fake != nil && fake.Parent != node {
				t.Fatal("fake operator acquired an artificial semantic mrow parent")
			}
		}
	}
}
