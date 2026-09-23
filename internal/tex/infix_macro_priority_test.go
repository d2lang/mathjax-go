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
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

// Keep every ordered attribute and own-property pair. The fixture-specific
// boundaries below retain complete accepted trees rather than dropping fields.
func infixMacroPairs(m *ordered.Map[mml.Property]) []map[string]any {
	out := []map[string]any{}
	m.Range(func(name string, value any) bool {
		if mml.IsInherit(value) {
			value = "_inherit_"
		}
		out = append(out, map[string]any{"name": name, "value": value})
		return true
	})
	return out
}
func infixMacroTree(t *testing.T, n *mml.Node) map[string]any {
	t.Helper()
	children := []any{}
	for _, child := range n.Children {
		if child.Parent != n {
			t.Fatal("child parent changed")
		}
		children = append(children, infixMacroTree(t, child))
	}
	kind := n.Kind
	if kind == "mrow" && n.Flags.Inferred {
		kind = "inferredMrow"
	}
	var text any
	if n.Kind == "text" {
		text = n.Text
	}
	return map[string]any{"kind": kind, "text": text, "attributes": map[string]any{"explicit": infixMacroPairs(n.Attributes.Explicit()), "inherited": infixMacroPairs(n.Attributes.Inherited()), "defaults": infixMacroPairs(n.Attributes.Defaults()), "global": infixMacroPairs(n.Attributes.Globals())}, "properties": infixMacroPairs(n.Properties), "children": children}
}
func TestInfixMacroPriorityPinnedReferences(t *testing.T) {
	type registration struct {
		Name, Body string
		Arguments  int
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, Scope, SVGSHA256 string
			Display                     bool
			Registration                *registration
			Tree                        any
		}
	}
	var limits struct {
		Baseline   string
		Boundaries map[string]struct {
			TeX, Reason, AcceptedReferenceCase, PrimarySVG, ExpectedSVG string
			Display                                                     bool
			Tree                                                        any
		}
	}
	for file, target := range map[string]any{"infix_macro_priority_mathjax_3_2_2.json": &fixture, "infix_macro_priority_boundaries.json": &limits} {
		b, err := os.ReadFile(filepath.Join("../../testdata", file))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 82 || limits.Baseline != "b8b28469b0c6357a93048864537bbbc61c56be56" || len(limits.Boundaries) != 22 {
		t.Fatal("unbound infix priority references")
	}
	expectedBoundaries := map[string]string{
		"over-public-operator-inline":     "ordinary declared-operator property insertion order",
		"over-public-operator-display":    "ordinary declared-operator property insertion order",
		"atop-public-operator-inline":     "ordinary declared-operator property insertion order",
		"atop-public-operator-display":    "ordinary declared-operator property insertion order",
		"atop-ordinary-inline":            "unchanged atop linethickness string type",
		"atop-ordinary-display":           "unchanged atop linethickness string type",
		"atop-true-alias-inline":          "unchanged atop linethickness string type",
		"atop-true-alias-display":         "unchanged atop linethickness string type",
		"above-public-operator-inline":    "ordinary declared-operator property insertion order",
		"above-public-operator-display":   "ordinary declared-operator property insertion order",
		"choose-public-operator-inline":   "ordinary declared-operator property insertion order",
		"choose-public-operator-display":  "ordinary declared-operator property insertion order",
		"brace-public-operator-inline":    "ordinary declared-operator property insertion order",
		"brace-public-operator-display":   "ordinary declared-operator property insertion order",
		"brack-public-operator-inline":    "ordinary declared-operator property insertion order",
		"brack-public-operator-display":   "ordinary declared-operator property insertion order",
		"public-operator-control-inline":  "ordinary declared-operator property insertion order",
		"public-operator-control-display": "ordinary declared-operator property insertion order",
		"choose-empty-override-inline":    "unchanged ordinary-prime metadata",
		"choose-empty-override-display":   "unchanged ordinary-prime metadata",
		"plain-prime-control-inline":      "unchanged ordinary-prime metadata",
		"plain-prime-control-display":     "unchanged ordinary-prime metadata",
	}
	expectedErrors := map[string]string{"spaced-infix-alias-inline": "AmbiguousUseOf", "spaced-infix-alias-display": "AmbiguousUseOf", "above-missing-second-argument-inline": "MissingArgFor", "above-missing-second-argument-display": "MissingArgFor", "over-missing-argument-inline": "MissingArgFor", "over-missing-argument-display": "MissingArgFor", "joined-infix-alias-inline": "AmbiguousUseOf", "joined-infix-alias-display": "AmbiguousUseOf"}
	raw, qualified := 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.Tree
			if b, ok := limits.Boundaries[c.Name]; ok {
				if expectedBoundaries[c.Name] != b.Reason || b.TeX != c.TeX || b.Display != c.Display || b.PrimarySVG != c.SVGSHA256 {
					t.Fatal("unbound complete boundary")
				}
				wantSVG, wantTree = b.ExpectedSVG, b.Tree
				qualified++
			} else {
				raw++
			}
			var root *mml.Node
			if c.Registration == nil {
				var err error
				root, err = NewCompiler().Compile(c.TeX, c.Display)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				state := newParseState()
				r := c.Registration
				definition := macroDefinition{body: r.Body, arguments: r.Arguments}
				state.macros[r.Name] = definition
				p := &parser{source: c.TeX, state: state, display: c.Display}
				children, stop, err := p.parseRow(0, false)
				if stop != "" {
					t.Fatal("unexpected parser stop", stop)
				}
				errorID := ""
				if err != nil {
					var typed *Error
					if !errors.As(err, &typed) {
						t.Fatal(err)
					}
					errorID = typed.ID
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
				if errorID != expectedErrors[c.Name] {
					t.Errorf("error ID %q; want %q", errorID, expectedErrors[c.Name])
				}
				if !reflect.DeepEqual(state.macros[r.Name], definition) {
					t.Error("registered definition mutated")
				}
			}
			b, err := json.Marshal(infixMacroTree(t, root))
			if err != nil {
				t.Fatal(err)
			}
			var tree any
			if err = json.Unmarshal(b, &tree); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tree, wantTree) {
				t.Error("complete ordered attributes/own-property tree differs")
			}
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			got, err := svg.NewTypesetter().Typeset(root, o)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != wantSVG {
				t.Errorf("complete SVG %s; want %s", h, wantSVG)
			}
			if out := os.Getenv("MATHJAX_INFIX_PRIORITY_EVIDENCE"); out != "" {
				if err = os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				for suffix, data := range map[string][]byte{".svg": []byte(got), ".json": b} {
					if err = os.WriteFile(filepath.Join(out, c.Name+suffix), data, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
	if raw != 60 || qualified != 22 {
		t.Fatal("reference classification changed", raw, qualified)
	}
}
