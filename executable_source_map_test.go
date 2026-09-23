// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

// Public TeX errors are rendered as merror SVG. These references check the
// complete public output and message; error IDs are retained observation metadata.
func TestExecutableSourceMapPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/executable_source_map_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG  string
			Display, Public bool
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 116 {
		t.Fatal("unbound executable source-map references")
	}
	count := 0
	for _, c := range fixture.Cases {
		if !c.Public {
			continue
		}
		count++
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
	if count != 112 {
		t.Fatalf("public references %d; want 112", count)
	}
}
