// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/tex"
	"os"
	"reflect"
	"testing"
)

func TestFramedInternalTextPinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *limitsTree
		}
	}
	data, err := os.ReadFile("testdata/framed_internal_text_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 64 {
		t.Fatal("unbound framed text references")
	}
	compiler := tex.NewCompiler()
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", got, c.SVGSHA256)
			}
			// Reusing this compiler across successful and failing content also checks
			// that an inner macro, font or color does not leak into the next compile.
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
			if !reflect.DeepEqual(got, c.Tree) {
				t.Errorf("complete explicit/own tree differs: %s", data)
			}
		})
	}
}
