// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	mathjax "github.com/d2lang/mathjax-go"
	"os"
	"testing"
)

func TestBraketBarMathJaxReference(t *testing.T) {
	data, err := os.ReadFile("testdata/braket_bar_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 18 {
		t.Fatalf("got %d cases, want 18", len(fixture.Cases))
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("whole SVG differs from pinned MathJax for %q", c.TeX)
			}
		})
	}
}
