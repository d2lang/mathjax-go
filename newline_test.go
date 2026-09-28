// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestNewlineOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/newline_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Counts           struct{ Total, Exact int }
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.Counts.Total != 9801 || len(fixture.Cases) != fixture.Counts.Exact {
		t.Fatal("unbound newline references")
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.Name == "" || c.SVG == "" || seen[c.Name] {
			t.Fatal("invalid newline reference", c.Name)
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

// These are explicitly qualified controls, not byte-exact parity claims for
// their unresolved source expressions. The original strings remain unmodified.
func TestNewlineQualifiedControls(t *testing.T) {
	data, err := os.ReadFile("testdata/newline_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		SVG   string `json:"svg"`
		Error string `json:"error"`
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, Qualification, LiteralTeX string
			Display                              bool
			Original, OriginalLiteral            result
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound newline controls")
	}
	safe, literal := 0, 0
	for _, c := range fixture.Cases {
		if c.Qualification == "safe-XML error attribute" {
			safe++
			t.Run(c.Name+"/safe-attribute", func(t *testing.T) {
				// Original MathJax serializes this one bare ampersand in an attribute.
				// Go must retain safe XML while keeping its message and rendering exact.
				const bare = `data-mjx-error="Misplaced &"`
				if !strings.Contains(c.Original.SVG, bare) {
					t.Fatal("missing original bare ampersand")
				}
				want := strings.Replace(c.Original.SVG, bare, `data-mjx-error="Misplaced &amp;"`, 1)
				options := mathjax.DefaultOptions()
				options.Display = c.Display
				got, err := mathjax.RenderWithOptions(c.TeX, options)
				if err != nil || got != want {
					t.Fatalf("safe attribute control differs: %v\ngot: %s\nwant: %s", err, got, want)
				}
			})
		}
		if c.LiteralTeX != "" {
			literal++
			t.Run(c.Name+"/literal-owner", func(t *testing.T) {
				if c.Original.SVG == "" || c.Original.SVG != c.OriginalLiteral.SVG || c.Original.Error != c.OriginalLiteral.Error {
					t.Fatal("literal lacks complete-original equivalence")
				}
				options := mathjax.DefaultOptions()
				options.Display = c.Display
				got, err := mathjax.RenderWithOptions(c.LiteralTeX, options)
				if err != nil || got != c.OriginalLiteral.SVG {
					t.Fatalf("literal owner control differs: %v\ngot: %s\nwant: %s", err, got, c.OriginalLiteral.SVG)
				}
			})
		}
	}
	if safe != 72 || literal != 74 {
		t.Fatalf("qualified inventory: safe=%d literal=%d", safe, literal)
	}
}
