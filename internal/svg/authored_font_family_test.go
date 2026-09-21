// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestAuthoredFontFamilyPinnedSVG(t *testing.T) {
	data, err := os.ReadFile("../../testdata/authored_font_family_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 78 {
		t.Fatal("incomplete authored font-family reference matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			before, nodes, parents := msInputSnapshot(root)
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			svg, err := NewTypesetter().Typeset(root, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", got, c.SVGSHA256)
			}
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("renderer changed original MathML attributes or content")
			}
			for i := range nodes {
				if nodes[i] != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("renderer replaced original node or parent identity")
				}
			}
			repeated, err := NewTypesetter().Typeset(root, opts)
			if err != nil || repeated != svg {
				t.Fatalf("repeat rendering changed font output: %v", err)
			}
			cloned, err := NewTypesetter().Typeset(root.Clone(), opts)
			if err != nil || cloned != svg {
				t.Fatalf("cloned rendering changed font output: %v", err)
			}
		})
	}
}
