// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// TestCombineRelationsPinnedMethods replays the actual MathJax 3.2.2 method
// observations, including creation-list order and detached-node identities.
// The complete before graph must roundtrip before any filtering is performed.
func TestCombineRelationsPinnedMethods(t *testing.T) {
	data, err := os.ReadFile("testdata/combine_relations_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []map[string]any `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 58 {
		t.Fatalf("method cases = %d, want 58", len(fixture.Cases))
	}
	seen := make(map[string]bool)
	for _, f := range fixture.Cases {
		name := f["spec"].(map[string]any)["name"].(string)
		if seen[name] {
			t.Fatalf("duplicate method case %q", name)
		}
		seen[name] = true
		t.Run(name, func(t *testing.T) {
			before := f["before"].(map[string]any)
			after := f["after"].(map[string]any)
			nodes := map[string]*mml.Node{}
			defs := map[string]map[string]any{}
			names := map[*mml.Node]string{}
			decode := func(v any) any {
				if x, ok := v.(map[string]any); ok && x["$type"] == "undefined" {
					return relationUndefined{}
				}
				if v == "_inherit_" {
					return mml.Inherit
				}
				return v
			}
			pairs := func(v any) *ordered.Map[any] {
				m := ordered.New[any]()
				for _, item := range v.([]any) {
					x := item.(map[string]any)
					m.Set(x["name"].(string), decode(x["value"]))
				}
				return m
			}
			var build func(map[string]any) *mml.Node
			build = func(d map[string]any) *mml.Node {
				id := d["id"].(string)
				if n := nodes[id]; n != nil {
					return n
				}
				kind := d["kind"].(string)
				inferred := kind == "inferredMrow"
				if inferred {
					kind = "mrow"
				}
				n := mml.NewNode(kind, nil, nil)
				if def, ok := texMMLFactory.Definition(kind); ok {
					n.Flags = def.Flags
				}
				n.Flags.Inferred = inferred
				n.Flags.NotParent = inferred
				nodes[id] = n
				names[n] = id
				defs[id] = d
				if text, ok := d["text"].(string); ok {
					n.Text = text
				}
				if tc, ok := d["texClass"].(float64); ok {
					n.TeXClass = mml.TeXClass(tc)
				}
				if attrs, ok := d["attributes"].(map[string]any); ok {
					n.Attributes = mml.NewAttributes(pairs(attrs["defaults"]), pairs(attrs["global"]))
					n.Attributes.SetList(pairs(attrs["explicit"]))
					pairs(attrs["inherited"]).Range(func(k string, v any) bool { n.Attributes.SetInherited(k, v); return true })
				}
				n.Properties = pairs(d["properties"])
				for _, child := range d["children"].([]any) {
					n.AppendChild(build(child.(map[string]any)))
				}
				refreshDynamicFlags(n)
				return n
			}
			root := build(before["tree"].(map[string]any))
			for _, x := range before["tracked"].([]any) {
				build(x.(map[string]any))
			}
			// Restore every saved parent identity, including detached nodes; construction
			// above is only graph decoding, never a new expected policy.
			for id, d := range defs {
				if parent, ok := d["parent"].(string); ok {
					nodes[id].Parent = nodes[parent]
				} else {
					nodes[id].Parent = nil
				}
			}
			ops := []*mml.Node{}
			for _, id := range before["registered"].([]any) {
				ops = append(ops, nodes[id.(string)])
			}
			encode := func(v any) any {
				if _, ok := v.(relationUndefined); ok {
					return map[string]any{"$type": "undefined"}
				}
				if mml.IsInherit(v) {
					return "_inherit_"
				}
				return v
			}
			outPairs := func(m *ordered.Map[any]) []any {
				r := []any{}
				m.Range(func(k string, v any) bool { r = append(r, map[string]any{"name": k, "value": encode(v)}); return true })
				return r
			}
			var tree func(*mml.Node, bool) map[string]any
			tree = func(n *mml.Node, plain bool) map[string]any {
				d := defs[names[n]]
				var attrs any
				if d["attributes"] != nil {
					attrs = map[string]any{"explicit": outPairs(n.Attributes.Explicit()), "inherited": outPairs(n.Attributes.Inherited()), "defaults": outPairs(n.Attributes.Defaults()), "global": outPairs(n.Attributes.Globals())}
				}
				var text any
				if n.Kind == "text" {
					text = n.Text
				}
				children := []any{}
				for _, c := range n.Children {
					children = append(children, tree(c, plain))
				}
				kind := n.Kind
				if n.Flags.Inferred {
					kind = "inferredMrow"
				}
				var tc any = float64(n.TeXClass)
				if d["texClass"] == nil {
					if n.TeXClass != mml.TeXClassNone {
						t.Fatal("null class changed")
					}
					tc = nil
				}
				if _, ok := d["texClass"].(map[string]any); ok {
					if n.TeXClass != mml.TeXClassNone {
						t.Fatal("undefined class changed")
					}
					tc = d["texClass"]
				}
				r := map[string]any{"kind": kind, "text": text, "attributes": attrs, "properties": outPairs(n.Properties), "texClass": tc, "children": children}
				if !plain {
					var parent any
					if n.Parent != nil {
						parent = names[n.Parent]
					}
					r["id"] = names[n]
					r["parent"] = parent
				}
				return r
			}
			snapshot := func(list []*mml.Node) map[string]any {
				registered := []any{}
				live := []any{}
				preorder := []any{}
				for _, n := range list {
					registered = append(registered, names[n])
					for p := n; p != nil; p = p.Parent {
						if p == root {
							live = append(live, names[n])
							break
						}
					}
				}
				root.Walk(func(n *mml.Node) bool {
					if n.Kind == "mo" {
						preorder = append(preorder, names[n])
					}
					return true
				})
				tracked := []any{}
				for _, item := range before["tracked"].([]any) {
					d := item.(map[string]any)
					r := tree(nodes[d["id"].(string)], false)
					r["label"] = d["label"]
					tracked = append(tracked, r)
				}
				return map[string]any{"tree": tree(root, false), "plain": tree(root, true), "registered": registered, "liveRegistered": live, "preorder": preorder, "tracked": tracked}
			}

			// As in the original replay, roundtrip the snapshot through JSON
			// solely to use JSON's numeric representation for comparison.
			// No attributes, properties, list order, or identity fields are
			// projected away in either comparison.
			jsonValue := func(v any) any {
				data, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				var out any
				if err := json.Unmarshal(data, &out); err != nil {
					t.Fatal(err)
				}
				return out
			}
			compare := func(stage string, got, want any) {
				t.Helper()
				if !reflect.DeepEqual(got, want) {
					gotJSON, _ := json.MarshalIndent(got, "", "  ")
					wantJSON, _ := json.MarshalIndent(want, "", "  ")
					t.Fatalf("%s differs\ngot:\n%s\nwant:\n%s", stage, gotJSON, wantJSON)
				}
			}
			compare("primary-before graph reconstruction", jsonValue(snapshot(ops)), before)
			compare("complete actual-primary after graph", jsonValue(snapshot(combineRelations(root, ops))), after)
		})
	}
}
