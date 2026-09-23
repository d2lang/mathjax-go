// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"github.com/d2lang/mathjax-go/internal/tex"
)

type identifierState struct {
	Path, Text string
	TeXClass   mml.TeXClass
	Properties map[string]any
	Variant    any
}

func TestIdentifierOperatorPinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Identifiers          []identifierState
		}
	}
	data, err := os.ReadFile("testdata/identifier_operator_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 60 {
		t.Fatal("unbound primary references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			output, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(output))); got != c.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", got, c.SVGSHA256)
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			po := pipeline.DefaultOptions()
			po.Display = c.Display
			direct, err := svg.NewTypesetter().Typeset(root, po)
			if err != nil || direct != output {
				t.Fatal("direct and public output differ")
			}
			states := []identifierState{}
			var walk func(*mml.Node, string)
			walk = func(n *mml.Node, path string) {
				if n.Kind == "mi" {
					text := ""
					for _, child := range n.Children {
						text += child.Text
					}
					props := map[string]any{}
					n.Properties.Range(func(k string, v any) bool { props[k] = v; return true })
					variant, _ := n.Attributes.Get("mathvariant")
					states = append(states, identifierState{path, text, n.TeXClass, props, variant})
				}
				for i, child := range n.Children {
					walk(child, fmt.Sprintf("%s/%d", path, i))
				}
			}
			walk(root, "root")
			// Normalize JSON numeric representation only, not any source field.
			data, err := json.Marshal(states)
			if err != nil {
				t.Fatal(err)
			}
			var got []identifierState
			if err = json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.Identifiers) {
				t.Error("raw primary identifier state differs")
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("measurement %dx%d %v; want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
