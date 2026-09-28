// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

// The original fixture sets these same attributes in the TeX postfilter,
// before the output jax creates its wrappers. This exercises the row-side
// negative splice and absent numeric consumers unavailable through array TeX.
func TestEmptyTableAttributeOriginalSVGs(t *testing.T) {
	data, err := os.ReadFile("testdata/empty_table_attributes_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, SourceTeX string
			Display         bool
			TableAttributes map[string]any
			Original        struct{ SVG string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 66 {
		t.Fatal("unbound controlled original table attributes")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.SourceTeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			for _, table := range root.Find("mtable") {
				for name, value := range c.TableAttributes {
					table.Attributes.Set(name, value)
				}
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete source-controlled SVG differs\ngot: %s\nwant: %s", got, c.Original.SVG)
			}
		})
	}
}
