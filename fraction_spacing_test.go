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

func TestFractionOperatorSpacingFrozenSVG(t *testing.T) {
	data, err := os.ReadFile("testdata/fraction_spacing_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name, TeX, SHA256 string
			Display           bool
			Width, Height     int
		}
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = fixture.Display
			got, err := mathjax.RenderWithOptions(fixture.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != fixture.SHA256 {
				t.Fatalf("complete SVG differs from frozen MathJax 3.2.2: got %s, want %s", hash, fixture.SHA256)
			}
			if fixture.Display {
				width, height, err := mathjax.Measure(fixture.TeX)
				if err != nil {
					t.Fatal(err)
				}
				if width != fixture.Width || height != fixture.Height {
					t.Fatalf("Measure = %dx%d, want %dx%d", width, height, fixture.Width, fixture.Height)
				}
			}
		})
	}
}
