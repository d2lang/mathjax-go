// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestNonscriptOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/nonscript_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2032 {
		t.Fatal("unbound Nonscript references")
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid Nonscript reference", c.Name)
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

func TestNonscriptArrayCompositionOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/nonscript_array_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 572 {
		t.Fatal("unbound Nonscript array references")
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid Nonscript array reference", c.Name)
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

// The original filter can splice a required fraction, root, or accent child
// and then throw during conversion. The Go API returns an internal conversion
// error for those inputs rather than panicking or inventing rendered content.
func TestNonscriptRequiredChildConversionErrors(t *testing.T) {
	data, err := os.ReadFile("testdata/nonscript_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, Category string
			Display             bool
			Original            struct{ Error string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, c := range fixture.Cases {
		if c.Category != "original-runtime-conversion-error" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			if !strings.HasPrefix(c.Original.Error, "TypeError:") {
				t.Fatal("missing original conversion failure")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err == nil || !strings.Contains(err.Error(), "nonscript removal leaves ") || got != "" {
				t.Fatalf("expected bounded required-child conversion error, got %q, %v", got, err)
			}
		})
	}
	if count != 118 {
		t.Fatalf("unbound conversion failure references: %d", count)
	}
}
