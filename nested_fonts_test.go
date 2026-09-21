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
	"github.com/d2lang/mathjax-go/internal/tex"
)

type nestedFontTree struct {
	Kind       string            `json:"kind"`
	Text       *string           `json:"text"`
	Attributes map[string]any    `json:"attributes"`
	Children   []*nestedFontTree `json:"children"`
}

func projectNestedFont(n *mml.Node) *nestedFontTree {
	r := &nestedFontTree{Kind: n.Kind, Attributes: map[string]any{}, Children: []*nestedFontTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	for _, k := range n.Attributes.ExplicitNames() {
		r.Attributes[k], _ = n.Attributes.GetExplicit(k)
	}
	for _, c := range n.Children {
		r.Children = append(r.Children, projectNestedFont(c))
	}
	return r
}
func TestNestedFontsPinnedASTAndSVG(t *testing.T) {
	b, e := os.ReadFile("internal/tex/testdata/nested_fonts_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *nestedFontTree
		}
	}
	if e = json.Unmarshal(b, &fixture); e != nil {
		t.Fatal(e)
	}
	if len(fixture.Cases) != 56 {
		t.Fatal("incomplete nested font matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, e := tex.NewCompiler().Compile(c.TeX, c.Display)
			if e != nil {
				t.Fatal(e)
			}
			root.Walk(func(n *mml.Node) bool {
				if _, ok := n.Property("go-resolved-font-scope"); ok {
					t.Error("temporary font scope escaped compilation")
				}
				return true
			})
			got := projectNestedFont(root)
			encoded, _ := json.Marshal(got)
			if e = json.Unmarshal(encoded, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, c.Tree) {
				want, _ := json.Marshal(c.Tree)
				t.Errorf("AST got %s\nwant %s", encoded, want)
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, opts)
			if e != nil {
				t.Fatal(e)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); h != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", h, c.SVGSHA256)
			}
		})
	}
}
