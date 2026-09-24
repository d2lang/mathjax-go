// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestAdjacentRelationsPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/adjacent_relations_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, Diagnostic string
			Display                    bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 68 {
		t.Fatal("unbound adjacent-relation references")
	}
	diagnostics := map[string]string{}
	seen := map[string]bool{}
	asserted, retained := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" || c.Diagnostic != diagnostics[c.Name] {
			t.Fatal("unapproved adjacent-relation reference inventory", c.Name)
		}
		seen[c.Name] = true
		if c.Diagnostic != "" {
			retained++
			continue
		}
		asserted++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
	if asserted != 68 || retained != len(diagnostics) {
		t.Fatal("adjacent-relation reference scope changed", asserted, retained)
	}
}
