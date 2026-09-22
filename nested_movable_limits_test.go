// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestNestedMovableLimitsPublicPrimary(t *testing.T) {
	data, err := os.ReadFile("testdata/nested_movable_limits_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 40 {
		t.Fatal("incomplete public matrix")
	}
	size := regexp.MustCompile(`(?:width|height)="([\d.]+)ex"`)
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(got, `data-mml-node="merror"`) {
				t.Fatal("unexpected error output")
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != c.SVGSHA256 {
				t.Fatalf("complete primary SVG %s want %s", h, c.SVGSHA256)
			}
			dimensions := size.FindAllStringSubmatch(got, -1)
			if len(dimensions) != 2 {
				t.Fatal("invalid SVG dimensions")
			}
			for i, want := range []int{c.Width, c.Height} {
				v, err := strconv.ParseFloat(dimensions[i][1], 64)
				if err != nil || int(math.Ceil(v*8)) != want {
					t.Fatalf("dimension %d differs", i)
				}
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("Measure=%d,%d,%v", w, h, err)
				}
			}
		})
	}
}
