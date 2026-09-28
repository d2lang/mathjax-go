// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestCrampedStyleOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/cramped_style_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1332 {
		t.Fatal("unbound cramped style references")
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid cramped style reference", c.Name)
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
				t.Fatalf("complete original SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
}

// Keep the published raw receipt untouched while promoting its optional-style
// cases to complete original SVG assertions.
func TestCrampedStylePromotesPublishedStackReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/cramped_substack_residuals.json")
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
		t.Fatal("unbound published cramped-stack references")
	}
	promoted := 0
	for _, c := range fixture.Cases {
		if !strings.Contains(c.TeX, `\cramped[\scriptscriptstyle]`) {
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
	if promoted != 24 {
		t.Fatalf("promoted %d published references, want 24", promoted)
	}
}
