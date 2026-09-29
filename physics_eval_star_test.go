// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPhysicsEvalStarOriginals(t *testing.T) {
	data, err := os.ReadFile("testdata/physics_eval_star_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 212 {
		t.Fatal("unbound original Eval star controls")
	}
	names := map[string]bool{}
	pairs := map[struct {
		tex     string
		display bool
	}]bool{}
	for _, c := range fixture.Cases {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || names[c.Name] || pairs[key] {
			t.Fatal("invalid or duplicate Eval star original", c.Name)
		}
		names[c.Name], pairs[key] = true, true
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
