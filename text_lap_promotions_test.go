// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

// Promote the published observations without modifying their raw receipts or
// counting them as newly captured original references.
func TestTextLapPromotesPublishedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/mathlap_style_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX string
			Display   bool
			Original  struct{ SVG string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound published text-lap references")
	}
	promoted := 0
	for _, c := range fixture.Cases {
		if c.TeX != `A\clap{x $y^2$}B` {
			continue
		}
		promoted++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if c.Original.SVG == "" || got != c.Original.SVG {
				t.Fatalf("published original SVG differs\ngot: %s\nwant: %s", got, c.Original.SVG)
			}
		})
	}
	if promoted != 2 {
		t.Fatalf("promoted %d published references, want 2", promoted)
	}
}
