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

type primeWhitespaceTree struct {
	Kind                   string
	Text                   *string
	Attributes, Properties map[string]any
	Children               []*primeWhitespaceTree
}

func TestPrimeWhitespacePinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256, RuntimeError string
			Display                            bool
			Width, Height                      int
			PropertiesTree                     *primeWhitespaceTree
		}
	}
	var bounds struct {
		Baseline               string
		InheritedPrimeMetadata map[string][]struct {
			Path       []int
			Properties map[string]any
		}
		UnchangedNonPrime map[string]struct {
			SVGSHA256 string
			Tree      *primeWhitespaceTree
		}
	}
	for name, target := range map[string]any{
		"testdata/prime_whitespace_mathjax_3_2_2.json": &fixture,
		"testdata/prime_whitespace_boundaries.json":    &bounds,
	} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 136 || bounds.Baseline != "c499d398cdfa3a34b0dae88dd0dd9c205c261ee2" || len(bounds.InheritedPrimeMetadata) != 124 || len(bounds.UnchangedNonPrime) != 4 {
		t.Fatal("unbound whitespace references")
	}
	rawSVG, runtimeErrors := 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			again, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil || svg != again {
				t.Fatal("unstable repeated render")
			}
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			boundary, unchanged := bounds.UnchangedNonPrime[c.Name]
			if c.RuntimeError != "" {
				runtimeErrors++
				if c.RuntimeError != "TypeError: Cannot read properties of null (reading '4')" || c.SVGSHA256 != "" || c.PropertiesTree != nil {
					t.Fatal("changed primary crash boundary")
				}
				// The primary supplies no output oracle for NEL. Its method-level
				// cursor policy is tested separately; do not reproduce its crash.
				if !unchanged {
					return
				}
			}
			if unchanged {
				wantSVG, wantTree = boundary.SVGSHA256, boundary.Tree
			} else {
				rawSVG++
				// Preserve the inherited D060/D061 metadata boundary at exact
				// recorded paths. Explicit attributes and SVG remain raw primary.
				for _, q := range bounds.InheritedPrimeMetadata[c.Name] {
					n := wantTree
					for _, i := range q.Path {
						if n == nil || i < 0 || i >= len(n.Children) {
							t.Fatal("missing prime path")
						}
						n = n.Children[i]
					}
					if n == nil || n.Kind != "mo" || !reflect.DeepEqual(n.Properties, q.Properties) || len(q.Properties) != 2 || q.Properties["variantForm"] != true {
						t.Fatal("changed inherited prime properties")
					}
					if _, ok := q.Properties["pseudoscript"].(bool); !ok {
						t.Fatal("invalid inherited pseudoscript")
					}
					delete(n.Properties, "pseudoscript")
				}
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != wantSVG {
				t.Errorf("complete SVG %s; want %s", got, wantSVG)
			}
			n, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(limitsProjection(n))
			if err != nil {
				t.Fatal(err)
			}
			var got *primeWhitespaceTree
			if err = json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wantTree) {
				t.Error("complete explicit and own-property tree differs")
			}
			if c.Display && !unchanged {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("measurement %dx%d %v; want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
	if rawSVG != 128 || runtimeErrors != 6 {
		t.Fatal("changed output/error coverage")
	}
}
