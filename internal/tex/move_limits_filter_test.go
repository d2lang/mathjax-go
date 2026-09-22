// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestMoveLimitsRegisteredOrderAndOwnership(t *testing.T) {
	b, err := os.ReadFile("testdata/limits_filter_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		MathjaxGitCommit string
		Outputs          []struct {
			Spec struct {
				Kinds      []string
				OuterFirst bool
			}
			Before []struct {
				Name, Kind, Parent string
				Children           []string
			}
			After []struct {
				Name, Kind  string
				Parent      *string
				Children    []string
				Replacement *struct {
					Kind, ID                               string
					AttributesShared, PropertyObjectShared bool
					Properties                             map[string]any
					Children, ChildParents                 []string
				}
			}
			Tree  map[string]any
			Lists map[string][]string
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Outputs) != 12 || f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("missing primary ordered controls")
	}
	for _, c := range f.Outputs {
		lists := map[string][]*mml.Node{}
		register := func(n *mml.Node) *mml.Node { lists[n.Kind] = append(lists[n.Kind], n); return n }
		mark := func(k string) []*mml.Node {
			a := []*mml.Node{limitToken("mi", "i")}
			if k == "munderover" {
				a = append(a, limitToken("mi", "n"))
			}
			return a
		}
		build := func(k string, base *mml.Node) *mml.Node {
			n := texMMLFactory.Create(k)
			if base != nil {
				n.SetChildren(append([]*mml.Node{base}, mark(k)...))
			}
			n.Attributes.Set("displaystyle", false)
			n.SetProperty("movablelimits", true)
			n.Attributes.Set("custom", "retained")
			return register(n)
		}
		base := limitToken("mi", "x")
		base.SetProperty("movablelimits", true)
		var inner, outer *mml.Node
		if c.Spec.OuterFirst {
			outer = build(c.Spec.Kinds[0], nil)
			inner = build(c.Spec.Kinds[1], base)
			outer.SetChildren(append([]*mml.Node{inner}, mark(c.Spec.Kinds[0])...))
		} else {
			inner = build(c.Spec.Kinds[1], base)
			outer = build(c.Spec.Kinds[0], inner)
		}
		for _, n := range []*mml.Node{inner, outer} {
			n.SetProperty("movesupsub", true)
			n.SetProperty("texprimestyle", true)
			n.SetProperty("unknownCopyProbe", "kept")
		}
		detached := build("mover", limitToken("mi", "q"))
		root := node("math", outer)
		old := map[string]*mml.Node{"outer": outer, "inner": inner, "base": base, "detached": detached}
		ids := map[*mml.Node]string{}
		for _, a := range c.Before {
			n := old[a.Name]
			ids[n] = a.Name
			if len(a.Children) != len(n.Children) {
				t.Fatal("setup children")
			}
			for i, child := range n.Children {
				ids[child] = a.Children[i]
			}
			if n.Parent != nil {
				ids[n.Parent] = a.Parent
			}
		}
		collectorInput := root.Clone()
		result := moveMathLimitsFromLists(root, lists)
		collected := moveMathLimits(collectorInput)
		replacements := map[string]*mml.Node{}
		result.Walk(func(n *mml.Node) bool {
			for name, o := range old {
				if n != o && n.Attributes == o.Attributes {
					replacements[name] = n
				}
			}
			return true
		})
		for _, a := range c.After {
			if a.Replacement != nil {
				n := replacements[a.Name]
				if n == nil {
					t.Fatal("missing replacement")
				}
				ids[n] = a.Replacement.ID
			}
		}
		for _, a := range c.After {
			n := old[a.Name]
			if a.Parent == nil {
				if n.Parent != nil {
					t.Fatal("old parent retained")
				}
			} else if ids[n.Parent] != *a.Parent {
				t.Fatal("old parent identity differs")
			}
			children := []string{}
			for _, child := range n.Children {
				children = append(children, ids[child])
			}
			if !reflect.DeepEqual(children, a.Children) {
				t.Fatalf("%v/%t %s old children %v want%v", c.Spec.Kinds, c.Spec.OuterFirst, a.Name, children, a.Children)
			}
			if a.Replacement != nil {
				got := replacements[a.Name]
				want := a.Replacement
				if got.Kind != want.Kind || got.Attributes != n.Attributes || got.Properties == n.Properties || !reflect.DeepEqual(ownMap(got), want.Properties) {
					t.Fatal("copyAttributes ownership or properties mismatch")
				}
				for i, child := range got.Children {
					if ids[child] != want.Children[i] || ids[child.Parent] != want.ChildParents[i] {
						t.Fatal("replacement child/parent identity differs")
					}
				}
			}
		}
		for kind, expected := range c.Lists {
			actual := []string{}
			for _, n := range lists[kind] {
				actual = append(actual, ids[n])
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatal("registered list removal differs")
			}
		}
		// Rebuilding this same input through the compiler collector must retain the
		// same live kind/attrs/properties/text topology; its historical detached
		// wrapper trace can differ from creation-list order.
		var projection func(*mml.Node) map[string]any
		projection = func(n *mml.Node) map[string]any {
			children := []any{}
			for _, child := range n.Children {
				children = append(children, projection(child))
			}
			kind := n.Kind
			if kind == "mrow" && n.Flags.Inferred {
				kind = "inferredMrow"
			}
			return map[string]any{"kind": kind, "attributes": explicitMap(n), "properties": ownMap(n), "children": children}
		}
		var strip func(map[string]any) map[string]any
		strip = func(n map[string]any) map[string]any {
			delete(n, "id")
			for _, child := range n["children"].([]any) {
				strip(child.(map[string]any))
			}
			return n
		}
		if !reflect.DeepEqual(projection(result), projection(collected)) {
			t.Fatal("compiler collection changed live topology")
		}
		var flags func(*mml.Node) []mml.Flags
		flags = func(n *mml.Node) []mml.Flags {
			out := []mml.Flags{n.Flags}
			for _, child := range n.Children {
				out = append(out, flags(child)...)
			}
			return out
		}
		if !reflect.DeepEqual(flags(result), flags(collected)) {
			t.Fatal("compiler collection changed dynamic flags")
		}
		if !reflect.DeepEqual(projection(result), strip(c.Tree)) {
			t.Fatalf("live primary topology mismatch: %v", c.Spec)
		}
	}
}

func TestMoveLimitsActualPredicateBoundaries(t *testing.T) {
	b, err := os.ReadFile("testdata/limits_filter_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Predicates []struct {
			Spec                                                                                 map[string]any
			Changed                                                                              bool
			Kind                                                                                 string
			Retained, ChildrenSame, AttributesShared, OldDetached, ChildParent, PropertiesShared bool
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Predicates) != 20 {
		t.Fatal("missing pinned predicate controls")
	}
	for _, c := range f.Predicates {
		t.Run(c.Spec["name"].(string), func(t *testing.T) {
			s := c.Spec
			coreKind := "mi"
			text := "x"
			if k, ok := s["coreKind"]; ok {
				coreKind = k.(string)
				text = "∑"
			}
			core := limitToken(coreKind, text)
			base := core
			if s["wrapped"] == true {
				base = node("mstyle", core)
			}
			if v, ok := s["own"]; ok {
				base.SetProperty("movablelimits", v)
			}
			if s["coreOwn"] == true {
				core.SetProperty("movablelimits", true)
			}
			if s["baseAttribute"] == true {
				base.Attributes.Set("movablelimits", true)
			}
			if v, ok := s["explicit"]; ok {
				core.Attributes.Set("movablelimits", v)
			}
			if v, ok := s["inherited"]; ok {
				core.Attributes.SetInherited("movablelimits", v)
			}
			mark := limitToken("mi", "a")
			old := node("mover", base, mark)
			old.Attributes.Set("custom", "kept")
			if v, ok := s["display"]; ok {
				old.Attributes.Set("displaystyle", v)
			}
			if v, ok := s["displayInherited"]; ok {
				old.Attributes.SetInherited("displaystyle", v)
			}
			root := node("math", old)
			if s["detached"] == true {
				root = node("math", limitToken("mi", "z"))
				old.Parent = nil
			}
			lists := map[string][]*mml.Node{"mover": {old}}
			result := moveMathLimitsFromLists(root, lists)
			now := result.Children[0].Children[0]
			if s["detached"] == true {
				now = old
			}
			retained := false
			for _, n := range lists["mover"] {
				retained = retained || n == old
			}
			if (now != old) != c.Changed || now.Kind != c.Kind || retained != c.Retained || (now.Children[0] == base && now.Children[1] == mark) != c.ChildrenSame || (now.Attributes == old.Attributes) != c.AttributesShared || (old.Parent == nil) != c.OldDetached || (base.Parent == now) != c.ChildParent || (now.Properties == old.Properties) != c.PropertiesShared {
				t.Fatalf("actual primary predicate/ownership mismatch: %+v", c)
			}
		})
	}
}
