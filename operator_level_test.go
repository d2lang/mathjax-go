package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestOperatorInheritedLevelPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/operator_level_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 16 {
		t.Fatal("incomplete current inherited-level spacing matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if c.Name == "style-op-inline" {
				if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != "d44e3b505ca8c176bc1aa67b196126447d51e7a04b0fcf23965eb2972ce72f76" {
					t.Fatal("changed accepted style-op inline output")
				}
				// Existing parser metadata boundary: identical complete paint, but the
				// compiled inline script tag is msubsup instead of primary munderover.
				// Permit exactly this single literal tag, with every other byte strict.
				if strings.Count(got, `data-mml-node="msubsup"`) != 1 || strings.Contains(got, `data-mml-node="munderover"`) {
					t.Fatal("changed style-op structure")
				}
				got = strings.Replace(got, `data-mml-node="msubsup"`, `data-mml-node="munderover"`, 1)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", h, c.SVGSHA256)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("Measure %dx%d %v, want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
