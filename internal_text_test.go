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

func TestInternalTextPinnedReferences(t *testing.T) {
	type output struct {
		SVGSHA256 string
		Tree      *limitsTree
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX string
			Display   bool
			output
		}
	}
	var bounds struct {
		Baseline         string
		UnchangedCallers map[string]output
	}
	for file, target := range map[string]any{
		"testdata/internal_text_mathjax_3_2_2.json": &fixture,
		"testdata/internal_text_boundaries.json":    &bounds,
	} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 90 || bounds.Baseline != "ab6c3c973d305288864a96a8bbab5d5471762c6b" || len(bounds.UnchangedCallers) != 18 {
		t.Fatal("unbound text-box references")
	}
	primary, unchanged := 0, 0
	compiler := tex.NewCompiler()
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want := c.output
			if boundary, ok := bounds.UnchangedCallers[c.Name]; ok {
				want = boundary
				unchanged++
			} else {
				primary++
			}
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != want.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", got, want.SVGSHA256)
			}
			// Reuse the compiler across successes and errors to expose leaked
			// inner parser state. Include every own property, not only geometry.
			n, err := compiler.Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(limitsProjection(n))
			if err != nil {
				t.Fatal(err)
			}
			var got *limitsTree
			if err = json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want.Tree) {
				t.Error("complete explicit and own-property tree differs")
			}
		})
	}
	if primary != 72 || unchanged != 18 {
		t.Fatal("changed primary/caller coverage")
	}
}
