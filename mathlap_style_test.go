// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	mathjax "github.com/d2lang/mathjax-go"
	"os"
	"testing"
)

func TestMathLapStyleOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/mathlap_style_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1822 {
		t.Fatal("unbound MathLap style references")
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.Name == "" || c.SVG == "" || seen[c.Name] {
			t.Fatal("invalid original reference", c.Name)
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
