// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	mathjax "github.com/d2lang/mathjax-go"
	"os"
	"testing"
)

func TestNumericScriptPublicRendering(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256      string
			Display, MeasureAvailable bool
			Width, Height             int
			Registration              *struct{ Name string }
		}
	}
	b, e := os.ReadFile("testdata/numeric_script_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	var boundaries struct {
		Cases map[string]struct{ Kind, CandidateSVGSHA256 string }
	}
	b, e = os.ReadFile("testdata/numeric_script_boundaries.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &boundaries); e != nil {
		t.Fatal(e)
	}
	if len(f.Cases) != 120 || f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound fixture")
	}
	count := 0
	for _, c := range f.Cases {
		if c.Registration != nil {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			want := c.SVGSHA256
			q, qualified := boundaries.Cases[c.Name]
			if qualified {
				want = q.CandidateSVGSHA256
			}
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil {
				t.Fatal(e)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); got != want {
				t.Fatalf("whole SVG %s want %s", got, want)
			}
			repeated, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil || repeated != s {
				t.Fatal("repeat changed")
			}
			if c.Display && (!qualified || q.Kind == "inherited-prime-pseudoscript") {
				w, h, e := mathjax.Measure(c.TeX)
				if c.MeasureAvailable {
					if e != nil || w != c.Width || h != c.Height {
						t.Fatalf("measure %dx%d,%v want %dx%d", w, h, e, c.Width, c.Height)
					}
				} else if e == nil || e.Error() != "mathjax-go: SVG dimensions not found" {
					t.Fatal("primary percentage-width contract changed", e)
				}
			}
		})
	}
	if count != 108 {
		t.Fatal("public/registered inventory changed", count)
	}
}
