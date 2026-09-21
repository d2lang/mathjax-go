// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

type msPair struct {
	Name  string
	Value any
}
type msTree struct {
	Kind                string
	Text                *string
	Attributes          struct{ Explicit, Inherited, Defaults, Global []msPair }
	Properties          []msPair
	Flags               mml.Flags
	TeXClass, PrevClass mml.TeXClass
	PrevLevel           int
	Children            []*msTree
}

func msMap(pairs []msPair) *ordered.Map[mml.Property] {
	m := ordered.New[mml.Property]()
	for _, pair := range pairs {
		value := pair.Value
		if s, ok := value.(string); ok && s == "_inherit_" {
			value = mml.Inherit
		}
		m.Set(pair.Name, value)
	}
	return m
}

func (s *msTree) node() *mml.Node {
	kind := s.Kind
	if kind == "inferredMrow" {
		kind = "mrow"
	}
	n := mml.NewNode(kind, msMap(s.Attributes.Defaults), msMap(s.Attributes.Global))
	if s.Text != nil {
		n.Text = *s.Text
	}
	n.Flags = s.Flags
	n.TeXClass, n.PrevClass, n.PrevLevel = s.TeXClass, s.PrevClass, s.PrevLevel
	n.Attributes.SetList(msMap(s.Attributes.Explicit))
	for _, p := range s.Attributes.Inherited {
		n.Attributes.SetInherited(p.Name, p.Value)
	}
	for _, p := range s.Properties {
		n.SetProperty(p.Name, p.Value)
	}
	children := make([]*mml.Node, len(s.Children))
	for i, c := range s.Children {
		children[i] = c.node()
	}
	n.SetChildren(children)
	return n
}

// The typesetter may prepare TeX spacing state, but must never append quote
// nodes, rewrite text or change author attributes in the input MathML tree.
func msInputSnapshot(root *mml.Node) (string, []*mml.Node, []*mml.Node) {
	var nodes, parents []*mml.Node
	var rows []any
	root.Walk(func(n *mml.Node) bool {
		nodes = append(nodes, n)
		parents = append(parents, n.Parent)
		attributes := []any{}
		for _, layer := range []*ordered.Map[mml.Property]{n.Attributes.Explicit(), n.Attributes.Inherited(), n.Attributes.Defaults(), n.Attributes.Globals()} {
			pairs := []msPair{}
			layer.Range(func(k string, v any) bool { pairs = append(pairs, msPair{k, v}); return true })
			attributes = append(attributes, pairs)
		}
		rows = append(rows, []any{n.Kind, n.Text, attributes, len(n.Children)})
		return true
	})
	b, _ := json.Marshal(rows)
	return string(b), nodes, parents
}

func TestMsQuotesFrozenSameMathMLSVG(t *testing.T) {
	b, err := os.ReadFile("testdata/ms_quotes_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		AssetsSHA256 map[string]string
		Cases        []struct {
			Name, SHA256 string
			Display      bool
			Tree         *msTree
		}
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 60 {
		t.Fatal("incomplete frozen ms matrix")
	}
	if !reflect.DeepEqual(fixture.AssetsSHA256, map[string]string{
		"polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
		"mathjax.js":   "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
		"setup.js":     "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
	}) {
		t.Fatal("unpinned oracle assets")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			root := test.Tree.node()
			before, nodes, parents := msInputSnapshot(root)
			o := pipeline.DefaultOptions()
			o.Display = test.Display
			got, err := NewTypesetter().Typeset(root, o)
			if err != nil {
				t.Fatal(err)
			}
			if sha := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); sha != test.SHA256 {
				t.Errorf("complete SVG = %s, want pinned %s", sha, test.SHA256)
			}
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("renderer changed input MathML")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("renderer replaced original node or parent identity")
				}
			}
			repeat, err := NewTypesetter().Typeset(root, o)
			if err != nil || repeat != got {
				t.Fatalf("repeat rendering changed quotes: %v", err)
			}
			clone, err := NewTypesetter().Typeset(root.Clone(), o)
			if err != nil || clone != got {
				t.Fatalf("clone rendering changed quotes: %v", err)
			}
		})
	}
}
