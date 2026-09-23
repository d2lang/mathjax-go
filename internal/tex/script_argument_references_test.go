// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type argumentTree struct {
	Kind       string          `json:"kind"`
	Text       *string         `json:"text"`
	Attributes map[string]any  `json:"attributes"`
	Properties map[string]any  `json:"properties"`
	Children   []*argumentTree `json:"children"`
}

func argumentProjection(n *mml.Node) *argumentTree {
	r := &argumentTree{Kind: n.Kind, Attributes: map[string]any{}, Properties: map[string]any{}, Children: []*argumentTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		text := n.Text
		r.Text = &text
	}
	for _, k := range n.Attributes.ExplicitNames() {
		r.Attributes[k], _ = n.Attributes.GetExplicit(k)
	}
	for _, k := range n.Properties.Keys() {
		r.Properties[k], _ = n.Property(k)
	}
	for _, child := range n.Children {
		r.Children = append(r.Children, argumentProjection(child))
	}
	return r
}

func readArgumentJSON(t *testing.T, path string, into any) {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, into); e != nil {
		t.Fatal(e)
	}
}

func TestScriptArgumentPinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			PropertiesTree       *argumentTree
			Registration         *struct {
				Name, Body string
				Arguments  int
			}
		}
	}
	readArgumentJSON(t, "../../testdata/unbraced_prime_mathjax_3_2_2.json", &fixture)
	var boundaries struct {
		Baseline string
		Cases    map[string]struct {
			Kind, TeX, SVGSHA256 string
			Display              bool
			Tree                 *argumentTree
			Differences          []struct {
				Path              []int
				Field             string
				Primary, Accepted map[string]any
			}
		}
	}
	readArgumentJSON(t, "../../testdata/unbraced_prime_boundaries.json", &boundaries)
	var errorsFixture struct {
		MathjaxGitCommit string
		Rows             []struct {
			Name, Source, Remaining string
			Cursor                  int
			Error                   struct{ ID, Message string }
		}
	}
	readArgumentJSON(t, "testdata/script_argument_error_cursors.json", &errorsFixture)
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || errorsFixture.MathjaxGitCommit != fixture.MathjaxGitCommit || len(fixture.Cases) != 120 || len(errorsFixture.Rows) != 62 || len(boundaries.Cases) != 20 || boundaries.Baseline != "e93f5eb32042cd5a772f5b8fa3c6aabe540701da" {
		t.Fatal("unbound argument references")
	}
	registered, rawSVG, rawOwn, qualifiedOwn, errorsChecked := 0, 0, 0, 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			if c.Registration != nil {
				registered++
				r := c.Registration
				state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, parseErr := p.parseRow(0, false)
			for _, e := range errorsFixture.Rows {
				if e.Name == c.Name {
					errorsChecked++
					var pe *Error
					if !errors.As(parseErr, &pe) || pe.ID != e.Error.ID || pe.Message != e.Error.Message {
						t.Fatalf("typed error=%v want %s/%s", parseErr, e.Error.ID, e.Error.Message)
					}
					if p.source[p.pos:] != e.Remaining {
						t.Fatal("error consumed a different argument suffix")
					}
					if p.source != e.Source || len(utf16.Encode([]rune(p.source[:p.pos]))) != e.Cursor {
						t.Fatal("primary error cursor/source differs")
					}
				}
			}
			var root *mml.Node
			if parseErr != nil {
				var pe *Error
				if !errors.As(parseErr, &pe) {
					t.Fatal(parseErr)
				}
				root = mathError(pe.Message, c.Display)
			} else {
				if stop != "" {
					t.Fatal("unexpected stop", stop)
				}
				var e error
				children, e = p.amsTagFinalize(children)
				if e != nil {
					t.Fatal(e)
				}
				root = node("math", children...)
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
			}
			// Public cases additionally use the real Compiler, binding this registered
			// test-only finalization path to the production finalization sequence.
			if c.Registration == nil {
				actual, e := NewCompiler().Compile(c.TeX, c.Display)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(argumentProjection(root), argumentProjection(actual)) {
					t.Fatal("registered harness differs from public Compiler")
				}
				root = actual
			}
			root.Walk(func(n *mml.Node) bool {
				for _, ch := range n.Children {
					if ch.Parent != n {
						t.Fatal("child ownership changed")
					}
				}
				return true
			})
			wantTree, wantSVG := c.PropertiesTree, c.SVGSHA256
			b, qualified := boundaries.Cases[c.Name]
			if qualified {
				if b.TeX != c.TeX || b.Display != c.Display {
					t.Fatal("qualification input changed")
				}
				switch b.Kind {
				case "unchanged-output":
					if b.SVGSHA256 == c.SVGSHA256 || b.Tree == nil {
						t.Fatal("invalid unchanged output")
					}
					wantTree, wantSVG = b.Tree, b.SVGSHA256
				case "unchanged-pseudoscript-own-property":
					qualifiedOwn++
					rawSVG++
					for _, d := range b.Differences {
						at := wantTree
						for _, i := range d.Path {
							if i < 0 || i >= len(at.Children) {
								t.Fatal("invalid property path")
							}
							at = at.Children[i]
						}
						if d.Field != "properties" || !reflect.DeepEqual(at.Properties, d.Primary) {
							t.Fatal("primary own-property qualification changed")
						}
						if _, ok := d.Primary["pseudoscript"].(bool); !ok {
							t.Fatal("not the precise inherited boolean boundary")
						}
						delete(at.Properties, "pseudoscript")
						if !reflect.DeepEqual(at.Properties, d.Accepted) {
							t.Fatal("qualification changes another own property")
						}
					}
				default:
					t.Fatal("unknown qualification")
				}
			} else {
				rawSVG++
				rawOwn++
			}
			data, e := json.Marshal(argumentProjection(root))
			if e != nil {
				t.Fatal(e)
			}
			var got *argumentTree
			if e = json.Unmarshal(data, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, wantTree) {
				t.Error("complete explicit/own-property tree differs")
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, opts)
			if e != nil {
				t.Fatal(e)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(s))) != wantSVG {
				t.Error("complete SVG differs")
			}
		})
	}
	if registered != 20 || rawSVG != 118 || rawOwn != 100 || qualifiedOwn != 18 || errorsChecked != 62 {
		t.Fatal("reference scope changed", registered, rawSVG, rawOwn, qualifiedOwn, errorsChecked)
	}
}
