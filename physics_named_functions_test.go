// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPhysicsNamedFunctionReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/physics_named_functions_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		NamedFnMappings  map[string]string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		len(fixture.NamedFnMappings) != 29 || len(fixture.Cases) != 1000 {
		t.Fatal("unbound Physics NamedFn references")
	}
	for _, c := range fixture.Cases {
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
