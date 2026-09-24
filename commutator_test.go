// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestCommutatorPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/commutator_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
			Width, Height  int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 48 {
		t.Fatal("unbound commutator references")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" || c.Width <= 0 || c.Height <= 0 {
			t.Fatal("invalid commutator reference inventory", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Errorf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
			if c.Display {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Errorf("measure %dx%d, %v; want %dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
}
