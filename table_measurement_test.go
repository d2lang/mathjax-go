// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestTableMeasurementReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/table_measurement_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG  string
			Display         bool
			TableAttributes map[string]any
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 358 {
		t.Fatal("unbound table-measurement references")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid table-measurement reference inventory", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			var got string
			var err error
			if len(c.TableAttributes) == 0 {
				options := mathjax.DefaultOptions()
				options.Display = c.Display
				got, err = mathjax.RenderWithOptions(c.TeX, options)
			} else {
				// Match the original input-jax post-filter in the generator:
				// apply layout attributes to the completed MathML tree before
				// either implementation constructs output wrappers.
				root, compileErr := tex.NewCompiler().Compile(c.TeX, c.Display)
				if compileErr != nil {
					t.Fatal(compileErr)
				}
				tables := root.Find("mtable")
				if len(tables) == 0 {
					t.Fatal("layout control has no table")
				}
				for _, table := range tables {
					for name, value := range c.TableAttributes {
						table.Attributes.Set(name, value)
					}
				}
				options := pipeline.DefaultOptions()
				options.Display = c.Display
				got, err = svg.NewTypesetter().Typeset(root, options)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
}
