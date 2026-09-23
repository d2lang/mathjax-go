// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

// Use actual registered macros and the same successful/error finalization as
// Compiler.Compile. No public registration API or product parser is substituted.
func TestPodRegisteredMacroReferences(t *testing.T) {
	type output struct {
		SVGSHA256      string
		PropertiesTree *binomialTree
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX    string
			Display      bool
			Registration struct {
				Name, Body string
				Arguments  int
			}
			output
		}
	}
	data, err := os.ReadFile("../../testdata/pod_macros_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 20 {
		t.Fatal("unbound registered pod/pmod references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want := c.output
			state := newParseState()
			state.macros[c.Registration.Name] = macroDefinition{body: c.Registration.Body, arguments: c.Registration.Arguments}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, parseErr := p.parseRow(0, false)
			if stop != "" {
				t.Fatal("unexpected stop", stop)
			}
			var root *mml.Node
			if parseErr != nil {
				var typed *Error
				if !errors.As(parseErr, &typed) {
					t.Fatal(parseErr)
				}
				root = mathError(typed.Message, c.Display)
			} else {
				children, err = p.amsTagFinalize(children)
				if err != nil {
					t.Fatal(err)
				}
				root = node("math", children...)
				if c.Display {
					root.Attributes.Set("display", "block")
				}
				root.Walk(func(n *mml.Node) bool {
					for _, name := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
						n.RemoveProperty(name)
					}
					return true
				})
				setMathMLInheritance(root, c.Display)
				root = moveMathLimits(root)
				cleanMathMLAttributes(root)
			}
			root.Walk(func(n *mml.Node) bool {
				for _, child := range n.Children {
					if child.Parent != n {
						t.Fatal("child ownership lost")
					}
				}
				return true
			})
			data, err := json.Marshal(binomialProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var got *binomialTree
			if err = json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want.PropertiesTree) {
				t.Error("complete explicit-attribute and own-property tree differs")
			}
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			svg, err := svg.NewTypesetter().Typeset(root, o)
			if err != nil {
				t.Fatal(err)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); hash != want.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", hash, want.SVGSHA256)
			}
			if out := os.Getenv("MATHJAX_POD_MACRO_EVIDENCE"); out != "" {
				if err = os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				for name, body := range map[string][]byte{c.Name + ".svg": []byte(svg), c.Name + ".json": data} {
					if err = os.WriteFile(filepath.Join(out, name), body, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
	// Each compiler call restores the default registration after the explicit
	// zero-, one- and two-argument overrides and recursive error cases above.
	for _, source := range []string{`\pod{x}`, `\pmod{x}`} {
		root, err := NewCompiler().Compile(source, false)
		if err != nil || len(root.Find("merror")) != 0 || len(root.Find("mspace")) == 0 {
			t.Fatal("default registration lost", source, err)
		}
	}
}
