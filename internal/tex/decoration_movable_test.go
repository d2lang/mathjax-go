// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestOrdinaryDecorationNormalizesDirectBase(t *testing.T) {
	for _, command := range []string{"overline", "underline", "overrightarrow", "underrightarrow", "overleftarrow", "underleftarrow"} {
		for _, input := range []struct {
			name, source string
			normalized   bool
		}{
			{"sum", `{\sum}`, true}, {"product", `{\prod}`, true},
			{"mathop", `{\mathop{x}}`, true}, {"integral", `{\int}`, false},
			{"grouped-sum", `{{\sum}}`, false},
		} {
			t.Run(command+"/"+input.name, func(t *testing.T) {
				p := &parser{source: input.source, state: newParseState()}
				out, err := p.underOver(command)
				if err != nil || len(out) != 1 || len(out[0].Children) != 2 {
					t.Fatalf("ordinary constructor: %v", err)
				}
				annotation := out[0]
				if permission, ok := annotation.Property("subsupOK"); !ok || permission != true {
					t.Fatal("ordinary decoration lacks primary own script permission")
				}
				base, mark := annotation.Children[0], annotation.Children[1]
				if base.Parent != annotation || mark.Parent != annotation {
					t.Fatal("constructor child ownership changed")
				}
				value, present := base.Property("movablelimits")
				if input.normalized && (!present || value != false) {
					t.Fatalf("direct property = %v,%v; want own false", value, present)
				}
				if !input.normalized && present {
					t.Fatal("nonmovable or grouped direct base was mutated")
				}
				if input.name == "grouped-sum" {
					base.Walk(func(n *mml.Node) bool {
						if n.Kind == "mo" {
							if _, ok := n.Attributes.GetExplicit("movablelimits"); ok {
								t.Fatal("grouped descendant explicitly normalized")
							}
						}
						return true
					})
				}
			})
		}
	}
}

func TestDecorationNormalizationPreservesBaseIdentityAndMaps(t *testing.T) {
	// Exact direct-base mutation from ParseUtil.checkMovableLimits; the caller
	// must retain every unrelated property, attribute, child and parent.
	for _, kind := range []string{"mo", "mstyle", "TeXAtom", "mrow"} {
		t.Run(kind, func(t *testing.T) {
			base := node(kind, token("mi", "x"))
			base.Attributes.Set("mathcolor", "red")
			base.SetProperty("marker", "keep")
			base.SetProperty("movablelimits", true)
			parent := node("mrow", base, token("mi", "y"))
			child := base.Children[0]
			want := base.Clone()
			want.SetProperty("movablelimits", false)
			if kind == "mo" || kind == "mstyle" {
				want.Attributes.Set("movablelimits", false)
			}
			checkMovableLimits(base)
			if parent.Children[0] != base || base.Parent != parent || base.Children[0] != child || child.Parent != base {
				t.Fatal("base/child/parent identity changed")
			}
			if !reflect.DeepEqual(base.Attributes.Explicit(), want.Attributes.Explicit()) || !reflect.DeepEqual(base.Properties, want.Properties) {
				t.Fatal("complete own attribute/property maps differ")
			}
			beforeSecond := base.Clone()
			checkMovableLimits(base)
			if !reflect.DeepEqual(base.Attributes.Explicit(), beforeSecond.Attributes.Explicit()) || !reflect.DeepEqual(base.Properties, beforeSecond.Properties) || parent.Children[0] != base || base.Children[0] != child {
				t.Fatal("repeat normalization changed state or identity")
			}
		})
	}
}
