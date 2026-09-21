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

func TestOperatorInheritedLevelState(t *testing.T) {
	data, err := os.ReadFile("testdata/operator_level_state_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Returned                                    string
			Before, Previous, After, PreviousAfter, AfterNull *msTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 18 {
		t.Fatal("incomplete inherited-level boundary controls")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			node := c.Before.node()
			var previous *mml.Node
			if c.Previous != nil {
				previous = c.Previous.node()
			}
			before, nodes, parents := msInputSnapshot(node)
			returned := adjustTeXClass(node, previous)
			wantReturn := node
			if c.Returned == "previous" {
				wantReturn = previous
			} else if c.Returned != "node" {
				t.Fatal("unknown oracle return")
			}
			if returned != wantReturn {
				t.Fatal("changed return identity")
			}
			assertScriptClassState(t, node, c.After, "node")
			if previous != nil {
				assertScriptClassState(t, previous, c.PreviousAfter, "previous")
			}
			if c.AfterNull != nil {
				if adjustTeXClass(node, nil) != node {
					t.Fatal("null previous changed returned node")
				}
				assertScriptClassState(t, node, c.AfterNull, "after-null")
			}
			after, newNodes, newParents := msInputSnapshot(node)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("changed source attributes/tree")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("changed object/parent identity")
				}
			}
		})
	}
}

func TestOperatorInheritedLevelSamePrimaryMathML(t *testing.T) {
	data, err := os.ReadFile("../../testdata/operator_level_mathjax_3_2_2.json")
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
	if len(fixture.Cases) != 16 {
		t.Fatal("incomplete primary trees")
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
				t.Fatal("changed MathML source")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("changed object/parent identity")
				}
			}
			repeated, err := NewTypesetter().Typeset(root, options)
			if err != nil || repeated != got {
				t.Fatalf("repeat changed output: %v", err)
			}
		})
	}
}
