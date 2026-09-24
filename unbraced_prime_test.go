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

func TestUnbracedPrimePublicRendering(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256      string
			Display, MeasureAvailable bool
			Width, Height             int
			Registration              json.RawMessage
		}
	}
	b, e := os.ReadFile("testdata/unbraced_prime_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	var boundaries struct {
		Cases map[string]struct{ Kind, SVGSHA256 string }
	}
	b, e = os.ReadFile("testdata/unbraced_prime_boundaries.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &boundaries); e != nil {
		t.Fatal(e)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 120 {
		t.Fatal("unbound references")
	}
	count, promotedOperators := 0, 0
	for _, c := range f.Cases {
		if len(c.Registration) > 0 {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			want := c.SVGSHA256
			boundary := boundaries.Cases[c.Name]
			unchanged := boundary.Kind == "unchanged-output"
			if c.Name == "public-22-inline" || c.Name == "public-22-display" {
				if c.TeX != `\operatorname{’}` || c.Display != (c.Name == "public-22-display") || !unchanged || boundary.SVGSHA256 == c.SVGSHA256 {
					t.Fatal("changed original operator-name prime boundary")
				}
				// Complete child parsing now matches the untouched original output.
				// Keep the historical boundary record without selecting its old SVG.
				promotedOperators++
				unchanged = false
			}
			if unchanged {
				want = boundary.SVGSHA256
			}
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil {
				t.Fatal(e)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); got != want {
				t.Errorf("whole public SVG=%s want=%s", got, want)
			}
			r, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil || r != s {
				t.Fatal("repeat public render changed")
			}
			if c.Display && !unchanged {
				w, h, e := mathjax.Measure(c.TeX)
				if c.MeasureAvailable {
					if e != nil || w != c.Width || h != c.Height {
						t.Errorf("measurement=%dx%d,%v want=%dx%d", w, h, e, c.Width, c.Height)
					}
				} else if e == nil || e.Error() != "mathjax-go: SVG dimensions not found" {
					t.Fatal("percentage-width measurement contract changed", e)
				}
			}
		})
	}
	if count != 100 || promotedOperators != 2 {
		t.Fatal("public/registered scope changed", count)
	}
}
