// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type decorationSpec struct {
	Kind, Text             string
	Attributes, Properties map[string]any
	Children               []decorationSpec
}

func decorationInput(s decorationSpec) *mml.Node {
	if s.Kind == "text" {
		return mml.NewText(s.Text)
	}
	children := make([]*mml.Node, len(s.Children))
	for i, child := range s.Children {
		children[i] = decorationInput(child)
	}
	n := node(s.Kind, children...)
	// The registered TeXAtom constructor supplies its own ORD property.
	if s.Kind == "TeXAtom" {
		n.SetProperty("texClass", 0)
	}
	for key, value := range s.Attributes {
		n.Attributes.Set(key, value)
	}
	for key, value := range s.Properties {
		n.SetProperty(key, value)
	}
	refreshDynamicFlags(n)
	return n
}

func TestScriptedDecorationRegisteredMethod(t *testing.T) {
	data, err := os.ReadFile("../../testdata/scripted_decoration_registered_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Attached                                          bool
			Name                                              string
			Spec                                              decorationSpec
			Before, Normalized, PreviousBefore, PreviousAfter any
			CoreID                                            string
			Eligible, BaseIdentityPreserved, OuterPermission  bool
			Links                                             []struct {
				ID, Parent string
				Children   []string
			}
			NormalizedParent string
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 24 {
		t.Fatal("unbound registered cases")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			base := decorationInput(c.Spec)
			script := node("mo", mml.NewText("―"))
			script.Attributes.Set("accent", true)
			ids := map[*mml.Node]string{}
			old := []*mml.Node{}
			var mark func(*mml.Node, string)
			mark = func(n *mml.Node, id string) {
				ids[n] = id
				old = append(old, n)
				for i, child := range n.Children {
					mark(child, fmt.Sprintf("%s.%d", id, i))
				}
			}
			mark(base, "base")
			mark(script, "script")
			var previousParent *mml.Node
			if c.Attached {
				sibling := node("mi", mml.NewText("z"))
				sibling.Attributes.Set("mathcolor", "blue")
				previousParent = node("mrow", base, sibling)
				ids[previousParent] = "previous-parent"
				old = append(old, previousParent)
				mark(sibling, "sibling")
			}
			var project func(*mml.Node) any
			project = func(n *mml.Node) any {
				children := []any{}
				for _, child := range n.Children {
					children = append(children, project(child))
				}
				attrs, props := map[string]any{}, map[string]any{}
				n.Attributes.Explicit().Range(func(k string, v any) bool { attrs[k] = v; return true })
				n.Properties.Range(func(k string, v any) bool { props[k] = v; return true })
				var text any
				if n.Kind == "text" {
					text = n.Text
				}
				kind := n.Kind
				if n.Flags.Inferred {
					kind = "inferredMrow"
				}
				return map[string]any{"id": ids[n], "kind": kind, "text": text, "attributes": attrs, "properties": props, "embellished": n.Flags.Embellished, "spacelike": n.Flags.Spacelike, "children": children}
			}
			equal := func(actual, expected any, label string) {
				t.Helper()
				bytes, err := json.Marshal(actual)
				if err != nil {
					t.Fatal(err)
				}
				var roundtrip any
				if err = json.Unmarshal(bytes, &roundtrip); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(roundtrip, expected) {
					t.Fatalf("%s differs\nactual %s\nexpected %v", label, bytes, expected)
				}
			}
			equal(project(base), c.Before, "input registered class/flags/maps")
			if previousParent != nil {
				equal(project(previousParent), c.PreviousBefore, "previous parent before")
			}
			checkMovableLimits(base)
			normalized := normalizeDecorationBase(base)
			if normalized != base {
				ids[normalized] = "new-row"
				if len(normalized.Children) != 2 || normalized.Children[1] != base {
					t.Fatal("new row did not retain incoming base")
				}
				ids[normalized.Children[0]] = "new-empty"
			}
			if c.Eligible != (normalized != base) || !c.BaseIdentityPreserved || ids[limitsCore(base)] != c.CoreID {
				t.Fatal("family/core/identity differs")
			}
			// Compact Go mover is used only as an ownership host; the compared contract
			// is the unchanged registered helper's normalized base, not D071 permission.
			outer := node("mover", normalized, script)
			ids[outer] = "outer"
			equal(project(normalized), c.Normalized, "complete normalized base")
			if previousParent != nil {
				equal(project(previousParent), c.PreviousAfter, "previous parent and unrelated sibling after")
			}
			if ids[normalized.Parent] != c.NormalizedParent || len(old) != len(c.Links) {
				t.Fatal("ownership inventory differs")
			}
			for i, n := range old {
				want := c.Links[i]
				children := []string{}
				for _, child := range n.Children {
					children = append(children, ids[child])
				}
				if ids[n] != want.ID || ids[n.Parent] != want.Parent || !reflect.DeepEqual(children, want.Children) {
					t.Fatalf("original node/child identity changed at %s", want.ID)
				}
			}
			if _, ok := outer.Property("subsupOK"); ok {
				t.Fatal("separate D071 permission introduced")
			}
		})
	}
}

func TestOrdinaryScriptedDecorationCallers(t *testing.T) {
	for _, command := range []string{"overline", "underline", "overrightarrow", "underrightarrow", "overleftarrow", "underleftarrow"} {
		for _, display := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", command, display), func(t *testing.T) {
				p := &parser{source: `{\sum_i^n}`, display: display, state: newParseState()}
				out, err := p.underOver(command)
				if err != nil || len(out) != 1 {
					t.Fatalf("constructor: %v", err)
				}
				annotation := out[0]
				if _, ok := annotation.Property("subsupOK"); ok {
					t.Fatal("D071 permission introduced")
				}
				row := annotation.Children[0]
				if row.Kind != "mrow" || len(row.Children) != 2 || row.Children[0].Kind != "mo" || len(row.Children[0].Children) != 0 || row.Children[1].Kind != "munderover" {
					t.Fatal("parse-time family/zero-child row differs")
				}
				for _, key := range []string{"lspace", "rspace"} {
					if v, ok := limitsCore(row.Children[1]).Attributes.GetExplicit(key); !ok || v != 0 {
						t.Fatalf("core %s=%v,%t", key, v, ok)
					}
				}
				if row.Parent != annotation || row.Children[1].Parent != row || row.Children[0].Parent != row {
					t.Fatal("caller ownership differs")
				}
			})
		}
	}
}
