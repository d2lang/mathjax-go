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
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestAccentPolicyPublicASTAndSVG(t *testing.T) {
	b, err := os.ReadFile("testdata/accent_policy_public_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *nestedFontTree
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 36 {
		t.Fatal("missing public accent policy cases")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			got := projectNestedFont(root)
			encoded, _ := json.Marshal(got)
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			// These untouched underOver constructors already have explicit parent
			// attributes where the primary leaves them implicit. Only these four
			// fixture-specific attributes differ; their complete SVGs remain strict.
			data, _ := json.Marshal(c.Tree)
			var want *nestedFontTree
			if err = json.Unmarshal(data, &want); err != nil {
				t.Fatal(err)
			}
			key, kind := "", ""
			switch c.Name {
			case "overarrow-inline", "overarrow-display":
				key, kind = "accent", "mover"
			case "underarrow-inline", "underarrow-display":
				key, kind = "accentunder", "munder"
			}
			if key != "" {
				if len(want.Children) != 1 || len(want.Children[0].Children) != 1 {
					t.Fatal("changed arrow control tree")
				}
				n := want.Children[0].Children[0]
				if n.Kind != kind {
					t.Fatal("changed arrow control kind")
				}
				if _, ok := n.Attributes[key]; ok {
					t.Fatal("primary no longer has implicit arrow attribute")
				}
				n.Attributes[key] = true
			}
			if !reflect.DeepEqual(got, want) {
				expected, _ := json.Marshal(want)
				t.Errorf("AST got %s\nwant %s", encoded, expected)
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); h != c.SVGSHA256 {
				t.Errorf("complete SVG=%s want%s", h, c.SVGSHA256)
			}
			if c.Display {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Errorf("Measure=%dx%d,%v want%dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
}
