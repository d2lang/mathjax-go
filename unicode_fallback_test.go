// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestUnicodeFallbackPinnedSVG(t *testing.T) {
	data, err := os.ReadFile("testdata/unicode_fallback_mathjax_3_2_2.json")
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
	if len(fixture.Cases) != 54 {
		t.Fatal("incomplete Unicode fallback reference matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", got, c.SVGSHA256)
			}
		})
	}
}
