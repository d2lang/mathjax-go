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

func TestMathitFrozenSVG(t *testing.T) {
	data, err := os.ReadFile("testdata/mathit_mathjax_3_2_2.json")
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
	if len(fixtures.Cases) != 38 {
		t.Fatal("incomplete mathit reference matrix")
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
				t.Errorf("complete SVG = %s, want pinned MathJax %s", hash, fixture.SHA256)
			}
			if fixture.Display {
				w, h, err := mathjax.Measure(fixture.TeX)
				if err != nil || w != fixture.Width || h != fixture.Height {
					t.Errorf("Measure = %dx%d, %v; want %dx%d", w, h, err, fixture.Width, fixture.Height)
				}
			}
		})
	}
}
