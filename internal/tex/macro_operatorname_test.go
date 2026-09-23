// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

// D103 is a separate, preexisting constructor policy. These records retain
// every primary property and the precise accepted property arrays. No actual
// tree or SVG is normalized, and residual outputs are not primary passes.
func TestMacroBoundaryOperatorNameReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			macroReferenceCase
			PrimarySVGExact bool
			AcceptedDirect  *macroReferenceOutput
			MacroCount      int
			Properties      []struct {
				Path              []int
				Primary, Accepted []any
			}
		}
	}
	data, err := os.ReadFile("testdata/macro_operatorname_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 14 {
		t.Fatal("unbound primary operator-name records")
	}
	rawCount, residualCount := 0, 0
	for _, c := range fixture.Cases {
		if c.PrimarySVGExact {
			rawCount++
		} else {
			residualCount++
		}
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			for _, r := range c.Registrations {
				state.macros[r.Name] = macroReferenceDefinition(r)
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" {
				t.Fatalf("parse error %v, stop %q", err, stop)
			}
			if p.source != c.TeX || p.pos != len(c.TeX) || p.state != state || state.macroCount != c.MacroCount {
				t.Fatal("child expansion changed caller source/cursor/state or revisited old input")
			}
			root := node("math", children...)
			if c.Display {
				root.Attributes.Set("display", "block")
			}
			root.Walk(func(n *mml.Node) bool {
				for _, k := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
					n.RemoveProperty(k)
				}
				return true
			})
			setMathMLInheritance(root, c.Display)
			root = moveMathLimits(root)
			cleanMathMLAttributes(root)
			encoded, err := json.Marshal(macroReferenceTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var got, primary map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(c.Primary.Tree, &primary); err != nil {
				t.Fatal(err)
			}
			// Check every field, with only the named exact property-array pairs.
			// The primary tree remains untouched and available in the fixture.
			exceptions := map[string][2][]any{}
			for _, v := range c.Properties {
				path := ""
				for _, i := range v.Path {
					path += fmt.Sprintf("/children/%d", i)
				}
				path += "/properties"
				if _, exists := exceptions[path]; exists {
					t.Fatal("duplicate property boundary")
				}
				exceptions[path] = [2][]any{v.Primary, v.Accepted}
			}
			wantProperties := 2
			if c.Name == "direct-control-inline" || c.Name == "direct-control-display" {
				wantProperties = 1
			}
			if len(exceptions) != wantProperties {
				t.Fatal("property boundary inventory changed")
			}
			var compare func(any, any, string)
			compare = func(expected, actual any, path string) {
				if pair, ok := exceptions[path]; ok {
					if !reflect.DeepEqual(expected, pair[0]) || !reflect.DeepEqual(actual, pair[1]) || reflect.DeepEqual(pair[0], pair[1]) {
						t.Fatalf("precise D103 property pair differs at %s: primary=%#v pair=%#v actual=%#v", path, expected, pair, actual)
					}
					delete(exceptions, path)
					return
				}
				switch e := expected.(type) {
				case map[string]any:
					a, ok := actual.(map[string]any)
					if !ok || len(e) != len(a) {
						t.Fatalf("map differs at %s", path)
					}
					for k, v := range e {
						av, ok := a[k]
						if !ok {
							t.Fatalf("missing %s/%s", path, k)
						}
						compare(v, av, path+"/"+k)
					}
				case []any:
					a, ok := actual.([]any)
					if !ok || len(e) != len(a) {
						t.Fatalf("array differs at %s", path)
					}
					for i, v := range e {
						compare(v, a[i], fmt.Sprintf("%s/%d", path, i))
					}
				default:
					if !reflect.DeepEqual(expected, actual) {
						t.Fatalf("field differs at %s: %v vs %v", path, expected, actual)
					}
				}
			}
			compare(primary, got, "")
			if len(exceptions) != 0 {
				t.Fatal("unvisited exact property boundary")
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			rendered, err := svg.NewTypesetter().Typeset(root, opts)
			if err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("%x", sha256.Sum256([]byte(rendered)))
			if c.PrimarySVGExact {
				if c.AcceptedDirect != nil || digest != c.Primary.SVGSHA256 {
					t.Fatal("whole primary SVG changed")
				}
			} else {
				if c.AcceptedDirect == nil || digest == c.Primary.SVGSHA256 || digest != c.AcceptedDirect.SVGSHA256 {
					t.Fatal("D103 direct-route residual changed")
				}
				var direct any
				if err := json.Unmarshal(c.AcceptedDirect.Tree, &direct); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, direct) {
					t.Fatal("complete accepted direct-route tree changed")
				}
			}
		})
	}
	if rawCount != 6 || residualCount != 8 {
		t.Fatal("raw-primary/residual counts changed")
	}
}

func TestMacroBoundaryOperatorNameErrorOwnership(t *testing.T) {
	state := newParseState()
	state.macros["pick"] = macroDefinition{body: "#1", arguments: 1}
	source := `{a\pick} remaining`
	p := &parser{source: source, state: state}
	nodes, err := p.amsOperatorName("operatorname")
	macroReferenceError(t, err, &Error{ID: "MissingArgFor", Message: "Missing argument for \\pick"})
	if nodes != nil || p.source != source || p.pos != len(`{a\pick}`) || p.state != state || state.macroCount != 0 {
		t.Fatal("error path transferred child program or changed caller state")
	}
}
