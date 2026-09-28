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
	var historical struct {
		MathjaxGitCommit string
		Cases []struct {
			Name, TeX string
			Display bool
			Original struct { SVG, Error string }
		}
	}
	data, err := os.ReadFile("testdata/math_token_historical_boundaries.json")
	if err != nil { t.Fatal(err) }
	if err := json.Unmarshal(data, &historical); err != nil { t.Fatal(err) }
	if historical.MathjaxGitCommit != fixture.MathjaxGitCommit || len(historical.Cases) != 12 {
		t.Fatal("unbound historical whitespace originals")
	}

	rawSVG, runtimeErrors := 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, o)
			if c.RuntimeError != "" {
				runtimeErrors++
				if c.RuntimeError != "TypeError: Cannot read properties of null (reading '4')" || c.SVGSHA256 != "" || c.PropertiesTree != nil {
					t.Fatal("changed primary crash boundary")
				}
				bound := false
				for _, h := range historical.Cases {
					if h.Name == "prime-"+c.Name {
						bound = h.TeX == c.TeX && h.Display == c.Display && h.Original.SVG == "" && strings.Split(h.Original.Error, "\n")[0] == c.RuntimeError
					}
				}
				if !bound || err == nil || svg != "" || err.Error() != "no Unicode range for character U+0085" {
					t.Fatalf("original NEL runtime failure must remain a bounded error: %q, %v", svg, err)
				}
				again, repeated := mathjax.RenderWithOptions(c.TeX, o)
				if repeated == nil || again != "" || repeated.Error() != err.Error() {
					t.Fatal("unstable bounded runtime failure")
				}
				return
			}
			if err != nil { t.Fatal(err) }
			again, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil || svg != again { t.Fatal("unstable repeated render") }
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			if c.Name == "bom-no-prime-inline" || c.Name == "bom-no-prime-display" {
				bound := false
				for _, h := range historical.Cases {
					if h.Name == "prime-"+c.Name {
						bound = h.TeX == c.TeX && h.Display == c.Display && h.Original.Error == "" && h.Original.SVG == svg
					}
				}
				if !bound { t.Fatal("complete original BOM SVG differs") }
			}
			rawSVG++
			// Preserve historical metadata receipts while requiring the complete
			// original inherited properties, including the promoted BOM mi.
			for _, q := range bounds.InheritedPrimeMetadata[c.Name] {
				n := wantTree
				for _, i := range q.Path {
					if n == nil || i < 0 || i >= len(n.Children) { t.Fatal("missing prime path") }
					n = n.Children[i]
				}
				if n == nil || n.Kind != "mo" || !reflect.DeepEqual(n.Properties, q.Properties) || len(q.Properties) != 2 || q.Properties["variantForm"] != true { t.Fatal("changed inherited prime properties") }
				if _, ok := q.Properties["pseudoscript"].(bool); !ok { t.Fatal("invalid inherited pseudoscript") }
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
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("measurement %dx%d %v; want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
	if rawSVG != 130 || runtimeErrors != 6 {
		t.Fatal("changed output/error coverage")
	}
}
