// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestUnicodePrimePinnedReferences(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *stackTree
			PropertiesTree       *limitsTree
		}
	}
	data, err := os.ReadFile("testdata/unicode_prime_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var boundaries struct {
		Baseline          string
		UnchangedNonPrime map[string]struct {
			TeX, SVGSHA256, BaselineTreeSHA256 string
			Display                            bool
			Tree                               *limitsTree
		}
		InheritedPrimeMetadata map[string][]struct {
			Path         []int
			PrimaryValue bool
			PrimaryText  string
		}
	}
	data, err = os.ReadFile("testdata/unicode_prime_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &boundaries); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 118 || len(boundaries.UnchangedNonPrime) != 10 || len(boundaries.InheritedPrimeMetadata) != 86 || boundaries.Baseline != "3dcb09030b4404c4c5982a9cebfbe5c1b6e51cd9" {
		t.Fatal("unbound Unicode prime matrix")
	}
	primaryErrors := 0
	for _, c := range fixture.Cases {
		if c.Tree.Children[0].Children[0].Kind == "merror" {
			primaryErrors++
		}
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			boundary, unchanged := boundaries.UnchangedNonPrime[c.Name]
			if unchanged {
				if boundary.TeX != c.TeX || boundary.Display != c.Display || len(boundary.BaselineTreeSHA256) != 64 || boundary.SVGSHA256 == c.SVGSHA256 {
					t.Fatal("changed preexisting literal/non-prime boundary")
				}
				wantSVG, wantTree = boundary.SVGSHA256, boundary.Tree
			} else {
				for _, b := range boundaries.InheritedPrimeMetadata[c.Name] {
					n := limitsNodeAt(wantTree, b.Path)
					if n == nil || n.Kind != "mo" || n.Properties["pseudoscript"] != b.PrimaryValue || len(n.Children) != 1 || n.Children[0].Text == nil || *n.Children[0].Text != b.PrimaryText {
						t.Fatal("changed exact inherited metadata path")
					}
					delete(n.Properties, "pseudoscript")
				}
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "_texLimitsScriptOrigin") {
				t.Fatal("transient state escaped")
			}
			var got *limitsTree
			if err = json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wantTree) {
				t.Error("complete explicit/own-property tree differs")
			}
			if !unchanged {
				raw, _ = json.Marshal(projectStackTree(root))
				var explicit *stackTree
				_ = json.Unmarshal(raw, &explicit)
				if !reflect.DeepEqual(explicit, c.Tree) {
					t.Error("explicit AST must be raw primary")
				}
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != wantSVG {
				t.Errorf("whole SVG=%s want=%s", got, wantSVG)
			}
			repeated, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || repeated != svg {
				t.Fatal("repeat render changed")
			}
			if c.Display && !unchanged {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("measure=%dx%d,%v want=%dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
	if primaryErrors != 16 {
		t.Fatal("required primary errors changed")
	}
}
