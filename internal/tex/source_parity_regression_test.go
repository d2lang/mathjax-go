// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Sources: ts/core/MmlTree/MmlNodes/mathchoice.ts,
// ts/input/tex/{ParseMethods,ParseUtil}.ts,
// base/{BaseItems,BaseMappings,BaseMethods}.ts,
// ams/{AmsMappings,AmsMethods}.ts, and mathtools/MathtoolsMethods.ts.

package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestAMSEqnArrayExtendsDefinitionsForAboxed(t *testing.T) {
	root, err := NewCompiler().Compile(`\begin{align*}a&\Aboxed{=b}\end{align*}`, true)
	if err != nil {
		t.Fatal(err)
	}
	table := onlyAMSTable(t, root)
	for name, want := range map[string]any{
		"columnalign":   "right left right left",
		"columnspacing": "0em 2em 0em",
	} {
		if got, ok := table.Attributes.GetExplicit(name); !ok || got != want {
			t.Fatalf("%s = %#v, %v; want %#v", name, got, ok, want)
		}
	}
}

func TestGenfracUsesFixedFencePalette(t *testing.T) {
	tests := []struct {
		name, tex, size string
	}{
		{"display-style", `\genfrac{(}{)}{1pt}{0}{a}{b}`, "2.047em"},
		{"text-style", `\genfrac{(}{)}{1pt}{1}{a}{b}`, "1.2em"},
		{"inherited-display", `\genfrac{(}{)}{1pt}{}{a}{b}`, "2.047em"},
		{"inherited-script", `x_{\genfrac{(}{)}{1pt}{}{a}{b}}`, "1.2em"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(root.Find("MathChoice")) + len(root.Find("mathchoice")); got != 0 {
				t.Fatalf("unresolved MathChoice count = %d", got)
			}
			fences := make([]*mml.Node, 0, 2)
			for _, operator := range root.Find("mo") {
				if textContent(operator) == "(" || textContent(operator) == ")" {
					fences = append(fences, operator)
				}
			}
			if len(fences) != 2 {
				t.Fatalf("fixed fence count = %d, want 2", len(fences))
			}
			for index, fence := range fences {
				if got := fence.Attributes.ExplicitNames(); !reflect.DeepEqual(got, []string{"minsize", "maxsize"}) {
					t.Fatalf("fence %d explicit attrs = %v", index, got)
				}
				for _, name := range []string{"minsize", "maxsize"} {
					if got, ok := fence.Attributes.GetExplicit(name); !ok || got != test.size {
						t.Fatalf("fence %d %s = %#v, %v; want %q", index, name, got, ok, test.size)
					}
				}
				for name, want := range map[string]any{"fence": true, "stretchy": true, "symmetric": true} {
					if got, ok := fence.Attributes.Get(name); !ok || got != want {
						t.Fatalf("fence %d %s = %#v, %v; want %#v", index, name, got, ok, want)
					}
				}
				atom := fence.Parent
				if atom != nil && atom.Kind == "mrow" && atom.Flags.Inferred {
					atom = atom.Parent
				}
				wantClass := mml.TeXClassOpen
				if index == 1 {
					wantClass = mml.TeXClassClose
				}
				if atom == nil || atom.Kind != "TeXAtom" || atom.TeXClass != wantClass {
					t.Fatalf("fence %d atom = %#v, want class %d", index, atom, wantClass)
				}
			}
		})
	}
}

func TestMathChoiceResolvesByInheritedStyle(t *testing.T) {
	tests := []struct {
		name, tex string
		display   bool
		want      string
	}{
		{"display", `\mathchoice{D}{T}{S}{Q}`, true, "D"},
		{"text", `\mathchoice{D}{T}{S}{Q}`, false, "T"},
		{"script", `x_{\mathchoice{D}{T}{S}{Q}}`, true, "xS"},
		{"scriptscript", `x_{y_{\mathchoice{D}{T}{S}{Q}}}`, true, "xyQ"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, test.display)
			if err != nil {
				t.Fatal(err)
			}
			if got := textContent(root); got != test.want {
				t.Fatalf("text = %q, want %q", got, test.want)
			}
			if got := len(root.Find("MathChoice")) + len(root.Find("mathchoice")); got != 0 {
				t.Fatalf("unresolved MathChoice count = %d", got)
			}
		})
	}
}

func TestMathFontAndMathcharSourceVariants(t *testing.T) {
	root, err := NewCompiler().Compile(`S^{{\mathcal{W}}_\Lambda}`, true)
	if err != nil {
		t.Fatal(err)
	}
	variants := map[string]string{}
	for _, identifier := range root.Find("mi") {
		if variant, ok := identifier.Attributes.GetExplicit("mathvariant"); ok {
			variants[textContent(identifier)] = sourceValueString(variant)
		}
	}
	if got := variants["W"]; got != "-tex-calligraphic" {
		t.Fatalf("mathcal W variant = %q, want -tex-calligraphic", got)
	}
	if got := variants["Λ"]; got != "normal" {
		t.Fatalf("Lambda variant = %q, want normal", got)
	}

	greek, err := NewCompiler().Compile(`\alpha`, true)
	if err != nil {
		t.Fatal(err)
	}
	alpha := greek.Find("mi")[0]
	if got, _ := alpha.Attributes.Get("mathvariant"); got != "italic" {
		t.Fatalf("alpha mathvariant = %#v; want effective italic", got)
	}
	if got, ok := alpha.Attributes.GetExplicit("mathvariant"); ok {
		t.Fatalf("alpha explicit mathvariant = %#v; want cleaned inherited value", got)
	}
}
