// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestGenfracStylePrimaryReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/genfrac_style_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
			Merror         bool
			Style          *struct {
				DisplayStyle bool
				ScriptLevel  int
			}
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 46 {
		t.Fatal("unbound primary style references")
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
			// Public compile errors are rendered as merror. Check actual model
			// classification, rather than interpreting a nil outer error as success.
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil || root == nil {
				t.Fatalf("compile: root=%v error=%v", root, err)
			}
			hasError := false
			root.Walk(func(n *mml.Node) bool {
				hasError = hasError || n.Kind == "merror"
				return true
			})
			if hasError != c.Merror {
				t.Fatalf("rendered merror=%v; want %v", hasError, c.Merror)
			}
			outer := root
			for (outer.Kind == "math" || (outer.Kind == "mrow" && outer.Flags.Inferred)) && len(outer.Children) == 1 {
				outer = outer.Children[0]
			}
			if c.Style == nil {
				if outer.Kind == "mstyle" {
					t.Fatal("unexpected outer style wrapper")
				}
				return
			}
			if outer.Kind != "mstyle" || outer.Attributes == nil {
				t.Fatal("missing primary style wrapper")
			}
			display, displayOK := outer.Attributes.GetExplicit("displaystyle")
			level, levelOK := outer.Attributes.GetExplicit("scriptlevel")
			if !displayOK || !levelOK || display != c.Style.DisplayStyle || level != c.Style.ScriptLevel {
				t.Fatalf("style display=%v level=%v; want %+v", display, level, c.Style)
			}
		})
	}
}
