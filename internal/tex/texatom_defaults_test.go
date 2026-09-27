// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestTeXAtomDefaultClassAndProperty(t *testing.T) {
	n := node("TeXAtom", token("mi", "x"))
	property, ok := n.Property("texClass")
	if n.TeXClass != mml.TeXClassOrd || !ok || property != mml.TeXClassOrd {
		t.Fatalf("default actual/property class = %v/%v, want ORD/ORD", n.TeXClass, property)
	}
	n.SetProperty("texClass", mml.TeXClassVCenter)
	if n.TeXClass != mml.TeXClassOrd {
		t.Fatal("serialization property changed the actual class")
	}
	for _, class := range []mml.TeXClass{mml.TeXClassOrd, mml.TeXClassBin, mml.TeXClassVCenter, mml.TeXClassNone} {
		n := texAtom(token("mi", "x"), class)
		if property, _ := n.Property("texClass"); n.TeXClass != class || property != class {
			t.Fatalf("explicit class %v was not preserved", class)
		}
	}
}

// Unlike the deliberately restricted ID-anchor helper, this private fixture
// builder permits a serialization property. It never writes the actual field.
func texAtomPropertyBuild(t *testing.T, raw json.RawMessage) *mml.Node {
	t.Helper()
	var s struct {
		Kind, Text string
		Attributes map[string]any
		Properties map[string]int
		Children   []json.RawMessage
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	if s.Kind == "text" {
		return mml.NewText(s.Text)
	}
	children := make([]*mml.Node, len(s.Children))
	for i, c := range s.Children {
		children[i] = texAtomPropertyBuild(t, c)
	}
	n := node(s.Kind, children...)
	// The saved witness has at most one attribute on each node; reject any
	// extension that would need an ordered attribute decoder.
	if len(s.Attributes) > 1 {
		t.Fatal("unexpected multi-attribute witness")
	}
	for k, v := range s.Attributes {
		n.Attributes.Set(k, v)
	}
	for k, v := range s.Properties {
		if s.Kind != "TeXAtom" || k != "texClass" {
			t.Fatal("unexpected property")
		}
		n.SetProperty(k, mml.TeXClass(v))
	}
	return n
}

func TestTeXAtomPropertyOriginalSVG(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit, OriginalRecordSHA256, SVG, SVGSHA256 string
		Spec                                                   struct {
			Display bool
			Tree    json.RawMessage
		}
	}
	data, err := os.ReadFile("testdata/texatom_property_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		len(fixture.OriginalRecordSHA256) != 64 || fmt.Sprintf("%x", sha256.Sum256([]byte(fixture.SVG))) != fixture.SVGSHA256 {
		t.Fatal("unbound TeXAtom reference")
	}
	root := texAtomPropertyBuild(t, fixture.Spec.Tree)
	setMathMLInheritance(root, fixture.Spec.Display)
	options := pipeline.DefaultOptions()
	options.Display = fixture.Spec.Display
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.SVG {
		t.Errorf("property-only TeXAtom differs from original: got %x, want %s", sha256.Sum256([]byte(got)), fixture.SVGSHA256)
	}
}
