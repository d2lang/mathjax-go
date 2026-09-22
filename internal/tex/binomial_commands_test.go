// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type binomialTree struct {
	Kind       string          `json:"kind"`
	Text       *string         `json:"text"`
	Attributes map[string]any  `json:"attributes"`
	Properties map[string]any  `json:"properties"`
	Children   []*binomialTree `json:"children"`
}

func binomialProjection(n *mml.Node) *binomialTree {
	r := &binomialTree{Kind: n.Kind, Attributes: map[string]any{}, Properties: map[string]any{}, Children: []*binomialTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	for _, k := range n.Attributes.ExplicitNames() {
		r.Attributes[k], _ = n.Attributes.GetExplicit(k)
	}
	for _, k := range n.Properties.Keys() {
		r.Properties[k], _ = n.Property(k)
	}
	for _, c := range n.Children {
		r.Children = append(r.Children, binomialProjection(c))
	}
	return r
}

// Registered declarations use the existing parser macro table. Finalization below
// is the same successful fixed-package path as Compiler.Compile, with no new
// compiler registration API or parser/renderer substitution.
func TestBinomialRegisteredMacroReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Registration         struct {
				Name, Body string
				Arguments  int
			}
			PropertiesTree *binomialTree
		}
	}
	data, err := os.ReadFile("../../testdata/binomial_macros_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 24 {
		t.Fatal("unbound registered oracle")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			state.macros[c.Registration.Name] = macroDefinition{body: c.Registration.Body, arguments: c.Registration.Arguments}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" {
				t.Fatal(err, stop)
			}
			children, err = p.amsTagFinalize(children)
			if err != nil {
				t.Fatal(err)
			}
			root := node("math", children...)
			if c.Display {
				root.Attributes.Set("display", "block")
			}
			root.Walk(func(n *mml.Node) bool {
				n.RemoveProperty(resolvedFontScope)
				n.RemoveProperty(ambientFontSource)
				n.RemoveProperty(vectorFactoryToken)
				n.RemoveProperty(vectorFactoryDone)
				n.RemoveProperty(limitsScriptOrigin)
				return true
			})
			setMathMLInheritance(root, c.Display)
			root = moveMathLimits(root)
			cleanMathMLAttributes(root)
			root.Walk(func(n *mml.Node) bool {
				for _, ch := range n.Children {
					if ch.Parent != n {
						t.Fatal("child ownership lost")
					}
				}
				return true
			})
			data, err := json.Marshal(binomialProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var actual *binomialTree
			if err = json.Unmarshal(data, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, c.PropertiesTree) {
				t.Error("complete explicit and own-property tree differs")
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			actualSVG, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actualSVG))); got != c.SVGSHA256 {
				t.Errorf("whole SVG=%s want=%s", got, c.SVGSHA256)
			}
			if out := os.Getenv("MATHJAX_BINOMIAL_MACRO_EVIDENCE"); out != "" {
				if err = os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(out, c.Name+".json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(out, c.Name+".svg"), []byte(actualSVG), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	// One conversion's declaration cannot escape its fresh parser state.
	fresh, err := NewCompiler().Compile(`\binom{x}{y}`, false)
	if err != nil || len(fresh.Find("mfrac")) != 1 {
		t.Fatal("registered macro leaked to public compiler")
	}
}

func TestBinomialHandlerOwnershipAndStyle(t *testing.T) {
	for _, c := range []struct{ name, style string }{{"binom", ""}, {"dbinom", "D"}, {"tbinom", "T"}} {
		for _, display := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", c.name, display), func(t *testing.T) {
				p := &parser{source: `{\mmlToken{mi}[mathvariant=bold]{x}}{y}`, state: newParseState(), display: display}
				nodes, err := p.binomial(c.name, c.style)
				if err != nil {
					t.Fatal(err)
				}
				if len(nodes) != 1 {
					t.Fatal("handler node count")
				}
				root := nodes[0]
				row := root
				if c.style != "" {
					if root.Kind != "mstyle" || len(root.Children) != 1 {
						t.Fatal("style must enclose complete fence expression")
					}
					inferred := root.Children[0]
					if inferred.Kind != "mrow" || !inferred.Flags.Inferred || len(inferred.Children) != 1 {
						t.Fatal("mstyle inferred owner changed")
					}
					row = inferred.Children[0]
					if v, _ := root.Attributes.GetExplicit("displaystyle"); v != (c.style == "D") {
						t.Fatal("explicit style wrong", v)
					}
				}
				if row.Kind != "mrow" || row.Flags.Inferred || len(row.Children) != 3 || row.TeXClass != mml.TeXClassOrd {
					t.Fatal("not fixed ORD fence row")
				}
				for k, v := range map[string]any{"open": "(", "close": ")", "texClass": mml.TeXClassOrd} {
					if got, _ := row.Property(k); got != v {
						t.Fatal(k, got)
					}
				}
				f := row.Children[1]
				if f.Kind != "mfrac" || len(f.Children) != 2 {
					t.Fatal("fraction shape")
				}
				if v, _ := f.Attributes.GetExplicit("linethickness"); v != "0" {
					t.Fatal("Genfrac thickness must be the source string zero", v)
				}
				if v, _ := f.Property("withDelims"); v != true {
					t.Fatal("fixed fraction padding missing")
				}
				xs := f.Children[0].Find("mi")
				if len(xs) != 1 {
					t.Fatal("argument structure")
				}
				if v, _ := xs[0].Attributes.GetExplicit("mathvariant"); v != "bold" {
					t.Fatal("argument attributes lost")
				}
				if row.Children[0].Kind != "MathChoice" || row.Children[2].Kind != "MathChoice" {
					t.Fatal("fixed palettes must defer style selection")
				}
				root.Walk(func(n *mml.Node) bool {
					for _, ch := range n.Children {
						if ch.Parent != n {
							t.Fatal("child identity/parent lost")
						}
					}
					return true
				})
			})
		}
	}
}
