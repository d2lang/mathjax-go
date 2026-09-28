// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPhysicsBraceQuantityReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/physics_brace_quantity_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1626 {
		t.Fatal("unbound Physics brace-quantity references")
	}
	seenNames, seenInputs := make(map[string]bool), make(map[struct {
		tex     string
		display bool
	}]bool)
	for _, c := range fixture.Cases {
		input := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || seenNames[c.Name] || seenInputs[input] || c.SVG == "" {
			t.Fatal("invalid Physics brace-quantity inventory", c.Name)
		}
		seenNames[c.Name], seenInputs[input] = true, true
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
}
