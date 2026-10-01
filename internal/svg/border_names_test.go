// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestBorderNamesSourceStyles(t *testing.T) {
	data, err := os.ReadFile("../../testdata/border_names_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		StyleCases       []struct {
			Name, Value string
			Border      [4][3]string
		}
		SplitCases []struct {
			Name, Value string
			Parts       [4]string
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.StyleCases) != 263 || len(fixture.SplitCases) != 0 {
		t.Fatal("unbound source Styles references")
	}
	for _, c := range fixture.StyleCases {
		t.Run(c.Name, func(t *testing.T) {
			styles := parseWrapperStyles(c.Value)
			var got [4][3]string
			for i, side := range styles.border {
				got[i] = [3]string{side.width, side.style, side.color}
			}
			if got != c.Border {
				t.Fatalf("border components differ: got %q want %q", got, c.Border)
			}
		})
	}
	for _, c := range fixture.SplitCases {
		t.Run(c.Name, func(t *testing.T) {
			if got := splitTRBL(c.Value); got != c.Parts {
				t.Fatalf("source split differs for %q: got %q want %q", c.Value, got, c.Parts)
			}
		})
	}
}

func TestBorderNamesPreservesCompiledMathML(t *testing.T) {
	data, err := os.ReadFile("../../testdata/border_names_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			before, nodes, parents := msInputSnapshot(root)
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			for _, input := range []struct {
				name  string
				clone bool
			}{{"original", false}, {"repeat", false}, {"clone", true}} {
				candidate := root
				if input.clone {
					candidate = root.Clone()
				}
				got, err := NewTypesetter().Typeset(candidate, options)
				if err != nil {
					t.Fatal(err)
				}
				if got != c.SVG {
					t.Fatalf("%s rendering differs from primary SVG", input.name)
				}
			}
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("renderer changed compiled MathML")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("renderer replaced node or parent identity")
				}
			}
		})
	}
}
