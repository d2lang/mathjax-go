// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestArrayRowReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/array_rows_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2197 {
		t.Fatal("unbound array row references")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid array row inventory", c.Name)
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
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
}

// Both explicit blank rows and an empty final pending entry are valid input.
// Before the source finalization fix these public calls reached a negative
// EqnArray column slice and panicked instead of returning the original SVG.
func TestEmptyAlignedRowsDoNotPanic(t *testing.T) {
	for _, environment := range []string{"aligned", "align"} {
		for separators := 1; separators <= 2; separators++ {
			tex := `\begin{` + environment + `}` + strings.Repeat(`\\`, separators) + `\end{` + environment + `}`
			for _, display := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/display=%t", environment, separators, display), func(t *testing.T) {
					defer func() {
						if recovered := recover(); recovered != nil {
							t.Fatalf("valid blank-row table panicked: %v", recovered)
						}
					}()
					options := mathjax.DefaultOptions()
					options.Display = display
					svg, err := mathjax.RenderWithOptions(tex, options)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(svg, "data-mjx-error") || strings.Count(svg, `data-mml-node="mtr"`) != separators {
						t.Fatalf("explicit blank rows were not preserved: %s", svg)
					}
					if display {
						if _, _, err := mathjax.Measure(tex); err != nil {
							t.Fatal(err)
						}
					}
				})
			}
		}
	}
}
