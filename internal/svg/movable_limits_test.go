// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func predicateWrapper(n *mml.Node) *wrapper {
	w := &wrapper{node: n}
	for _, c := range n.Children {
		child := predicateWrapper(c)
		child.parent = w
		w.children = append(w.children, child)
	}
	return w
}

func TestMovableLimitCorePrimary(t *testing.T) {
	data, err := os.ReadFile("testdata/movable_limit_core_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, CoreKind                  string
			Display, Expected, WrapperValue bool
			FullTree                        *msTree
			CorePath                        []int
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 62 {
		t.Fatal("incomplete primary core matrix")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := c.FullTree.node()
			before, nodes, parents := msInputSnapshot(root)
			w := predicateWrapper(root).children[0].children[0]
			w.displayStyle = c.Display
			want := w.children[0]
			for _, i := range c.CorePath {
				want = want.children[i]
			}
			got := movableLimitCore(w.children[0])
			if got != want {
				t.Fatalf("core identity differs: got %v want %s", got, want.node.Kind)
			}
			wantKind := c.CoreKind
			if wantKind == "inferredMrow" {
				wantKind = "mrow"
			}
			if got.node.Kind != wantKind {
				t.Fatal("core kind differs")
			}
			if c.Expected != c.WrapperValue || w.hasMovableLimits() != c.Expected {
				t.Fatalf("predicate=%v primary=%v", w.hasMovableLimits(), c.Expected)
			}
			after, newNodes, newParents := msInputSnapshot(root)
			if after != before || len(nodes) != len(newNodes) {
				t.Fatal("predicate mutated source")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("predicate changed identity")
				}
			}
		})
	}
}

func TestNestedMovableLimitsSamePrimaryMathML(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nested_movable_limits_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, SVGSHA256 string
			Display         bool
			FullTree        *msTree
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 40 {
		t.Fatal("incomplete public primary matrix")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := c.FullTree.node()
			before, nodes, parents := msInputSnapshot(root)
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			got, err := NewTypesetter().Typeset(root, o)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != c.SVGSHA256 {
				t.Fatalf("whole same-MathML SVG %s want %s", h, c.SVGSHA256)
			}
			again, err := NewTypesetter().Typeset(root, o)
			if err != nil || again != got {
				t.Fatalf("repeat changed output: %v", err)
			}
			clone, err := NewTypesetter().Typeset(root.Clone(), o)
			if err != nil || clone != got {
				t.Fatalf("clone changed output: %v", err)
			}
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("render changed input")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("render changed identity")
				}
			}
		})
	}
}
