// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

// These four fixtures use actual registered Macro definitions at the compiler
// initialization seam. They do not imply a public macro-registration API.
func TestExecutableSourceMapRegisteredReferences(t *testing.T) {
	data, err := os.ReadFile("../../testdata/executable_source_map_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, ErrorID, ErrorMessage string
			Display, Public                       bool
			Registration                          *struct {
				Name, Body string
				Arguments  int
			}
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 130 {
		t.Fatal("unbound executable source-map references")
	}
	count := 0
	for _, c := range fixture.Cases {
		if c.Public {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			r := c.Registration
			if r == nil || r.Name == "" || r.Arguments != 0 {
				t.Fatal("unbound private Macro registration")
			}
			state := newParseState()
			state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			root, parseError, err := starMacroFixtureCompile(p)
			if err != nil {
				t.Fatal(err)
			}
			if c.ErrorID == "" {
				if parseError != nil {
					t.Errorf("parse error %v; want success", parseError)
				}
			} else if parseError == nil || parseError.ID != c.ErrorID || parseError.Message != c.ErrorMessage {
				t.Errorf("parse error %v; want %s: %s", parseError, c.ErrorID, c.ErrorMessage)
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
	if count != 4 {
		t.Fatalf("private references %d; want 4", count)
	}
}
