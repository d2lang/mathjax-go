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

func assertScriptClassState(t *testing.T, got *mml.Node, want *msTree, path string) {
	t.Helper()
	if got.TeXClass != want.TeXClass || got.PrevClass != want.PrevClass || got.PrevLevel != want.PrevLevel {
		t.Errorf("%s %s class/previous/level = %d/%d/%d, want %d/%d/%d", path, got.Kind, got.TeXClass, got.PrevClass, got.PrevLevel, want.TeXClass, want.PrevClass, want.PrevLevel)
	}
	if len(got.Children) != len(want.Children) {
		t.Fatalf("%s changed children", path)
	}
	for i := range got.Children {
		assertScriptClassState(t, got.Children[i], want.Children[i], fmt.Sprintf("%s/%d", path, i))
	}
}

func TestScriptClassImmediateTransfer(t *testing.T) {
	data, err := os.ReadFile("testdata/script_class_transfer_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name                                   string
			Before, Previous, After, PreviousAfter *msTree
			ReturnedPath                           []int
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 12 {
		t.Fatal("incomplete immediate-base transfer controls")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, previous := c.Before.node(), c.Previous.node()
			before, nodes, parents := msInputSnapshot(root)
			returned := setTeXClass(root, previous)
			expected := root
			for _, index := range c.ReturnedPath {
				expected = expected.Children[index]
			}
			if returned != expected {
				t.Fatal("changed setTeXclass return identity")
			}
			assertScriptClassState(t, root, c.After, "root")
			assertScriptClassState(t, previous, c.PreviousAfter, "previous")
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("transfer changed MathML input")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("changed node/parent identity")
				}
			}
		})
	}
}

func TestScriptClassSamePrimaryMathML(t *testing.T) {
	data, err := os.ReadFile("../../testdata/script_class_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, SVGSHA256     string
			Display             bool
			FullTree, AfterTree *msTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 18 {
		t.Fatal("incomplete registered primary trees")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := c.FullTree.node()
			before, nodes, parents := msInputSnapshot(root)
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != c.SVGSHA256 {
				t.Errorf("same-MathML complete SVG %s, want %s", h, c.SVGSHA256)
			}
			assertScriptClassState(t, root, c.AfterTree, "root")
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("render changed source MathML")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("render changed identity")
				}
			}
			repeat, err := NewTypesetter().Typeset(root, options)
			if err != nil || repeat != got {
				t.Fatalf("repeat changed spacing: %v", err)
			}
		})
	}
}
