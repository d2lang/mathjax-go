// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"math"
	"reflect"
	"testing"
)

func TestCheckMovableLimitsDirectBase(t *testing.T) {
	for _, kind := range []string{"mo", "mstyle", "TeXAtom", "mrow"} {
		t.Run(kind, func(t *testing.T) {
			base := node(kind, token("mi", "x"))
			base.SetProperty("movablelimits", true)
			child := base.Children[0]
			checkMovableLimits(base)
			if v, ok := base.Property("movablelimits"); !ok || v != false {
				t.Fatalf("direct property = %v,%v; want explicit false", v, ok)
			}
			v, ok := base.Attributes.GetExplicit("movablelimits")
			if want := kind == "mo" || kind == "mstyle"; ok != want || (want && v != false) {
				t.Fatalf("attribute = %v,%v; mo/mstyle only", v, ok)
			}
			if base.Children[0] != child || child.Parent != base || textContent(child) != "x" {
				t.Fatal("normalization replaced content")
			}
		})
	}
	for _, kind := range []string{"TeXAtom", "mrow", "msubsup", "munderover"} {
		t.Run("do-not-descend/"+kind, func(t *testing.T) {
			child := token("mo", "∑")
			child.SetProperty("movablelimits", true)
			base := node(kind, child)
			first := base.Children[0] // TeXAtom inserts its ordinary inferred row.
			before := base.Clone()
			checkMovableLimits(base)
			if !reflect.DeepEqual(base, before) || base.Children[0] != first {
				t.Fatal("grouped/scripted descendant changed")
			}
		})
	}
}

func TestCheckMovableLimitsPropertyTruthiness(t *testing.T) {
	for _, c := range []struct {
		name  string
		value any
		clear bool
	}{
		{"true", true, true}, {"false", false, false},
		{"nonempty-string", "false", true}, {"empty-string", "", false},
		{"one", 1, true}, {"zero", 0, false},
		{"int64-one", int64(1), true}, {"int64-zero", int64(0), false},
		{"fraction", 0.5, true}, {"negative-zero", math.Copysign(0, -1), false},
		{"null", nil, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			base := node("TeXAtom", token("mi", "x"))
			base.SetProperty("movablelimits", c.value)
			checkMovableLimits(base)
			want := c.value
			if c.clear {
				want = false
			}
			if got, ok := base.Property("movablelimits"); !ok || !reflect.DeepEqual(got, want) {
				t.Fatalf("property = %v,%v; want %v", got, ok, want)
			}
		})
	}
	base := node("TeXAtom")
	base.SetProperty("movablelimits", math.NaN())
	checkMovableLimits(base)
	if got, _ := base.Property("movablelimits"); !math.IsNaN(got.(float64)) {
		t.Fatal("NaN is falsy")
	}
}

func TestCheckMovableLimitsDictionaryAndAttributeBoundary(t *testing.T) {
	for _, text := range []string{"∑", "∏"} {
		base := token("mo", text)
		checkMovableLimits(base)
		if v, ok := base.Attributes.GetExplicit("movablelimits"); !ok || v != false {
			t.Fatalf("%s dictionary was not normalized", text)
		}
	}
	for _, explicit := range []bool{false, true} {
		base := token("mo", "∫")
		if explicit {
			base.Attributes.Set("movablelimits", true)
		}
		before := base.Clone()
		checkMovableLimits(base)
		if !reflect.DeepEqual(base, before) {
			t.Fatal("attribute-only or absent integral policy changed")
		}
	}
}

func TestCheckMovableLimitsContextualFormOrder(t *testing.T) {
	// A test-only operator distinguishes first matching dictionary entry from
	// searching all forms for a truthy flag, or honoring explicit form here.
	oldPrefix, oldInfix := mjOperatorDictionaryPrefix, mjOperatorDictionaryInfix
	t.Cleanup(func() { mjOperatorDictionaryPrefix, mjOperatorDictionaryInfix = oldPrefix, oldInfix })
	const symbol = "D058-test"
	mjOperatorDictionaryPrefix = append([]mjOperatorDictionaryEntry{{Operator: symbol, Definition: mjOperatorDefinition{Properties: mjOperatorProperties{{Name: "movablelimits", Value: true}}}}}, oldPrefix...)
	mjOperatorDictionaryInfix = append([]mjOperatorDictionaryEntry{{Operator: symbol, Definition: mjOperatorDefinition{}}}, oldInfix...)
	for _, first := range []bool{false, true} {
		base := token("mo", symbol)
		base.Attributes.Set("form", "prefix")
		if first {
			node("mrow", base, token("mi", "x"))
		} else {
			node("mrow", token("mi", "x"), base, token("mi", "y"))
		}
		checkMovableLimits(base)
		v, ok := base.Property("movablelimits")
		if ok != first || (first && v != false) {
			t.Fatalf("first=%v policy=%v,%v", first, v, ok)
		}
	}
}

func TestOverUnderSetDirectMovableProperty(t *testing.T) {
	for _, display := range []bool{false, true} {
		for _, source := range []string{`\overunderset{a}{b}{\sum}`, `\overunderset{a}{b}{\mathop{x}}`} {
			root, err := NewCompiler().Compile(source, display)
			if err != nil {
				t.Fatal(err)
			}
			base := root.Children[0].Children[0].Children[0]
			if base.Kind == "TeXAtom" {
				// Existing mathClass omits the primary's movablelimits property.
				// Preserve this separately recorded metadata boundary; changing
				// the constructor would exceed the direct-base caller correction.
				if v, ok := base.Property("movablelimits"); ok {
					t.Fatalf("mathop baseline metadata changed: %v", v)
				}
				continue
			}
			if v, ok := base.Property("movablelimits"); !ok || v != false {
				t.Fatalf("%s direct property=%v,%v", source, v, ok)
			}
		}
	}
}
