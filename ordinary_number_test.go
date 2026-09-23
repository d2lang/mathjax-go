// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	mathjax "github.com/d2lang/mathjax-go"
	"os"
	"strings"
	"testing"
)

func TestOrdinaryNumberPublicRendering(t *testing.T) {
	var f struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
		}
	}
	var b struct {
		Cases map[string]struct{ Status, BaselineSVG string }
	}
	for n, v := range map[string]any{"testdata/ordinary_number_mathjax_3_2_2.json": &f, "testdata/ordinary_number_boundaries.json": &b} {
		raw, e := os.ReadFile(n)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(raw, v); e != nil {
			t.Fatal(e)
		}
	}
	count := 0
	for _, c := range f.Cases {
		if !strings.HasPrefix(c.Name, "public-") {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			want := c.SVGSHA256
			if q, ok := b.Cases[c.Name]; ok && q.Status == "unsupported-primary-unchanged-Go" {
				if c.Name != "public-unicode-segmented-inline" && c.Name != "public-unicode-segmented-display" {
					t.Fatal("unbound missing primary output")
				}
				want = q.BaselineSVG
			}
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil {
				t.Fatal(e)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(s))) != want {
				t.Fatal("whole public SVG differs")
			}
			repeat, e := mathjax.RenderWithOptions(c.TeX, o)
			if e != nil || repeat != s {
				t.Fatal("repeat changed")
			}
		})
	}
	if count != 80 {
		t.Fatal("public inventory changed")
	}
}
