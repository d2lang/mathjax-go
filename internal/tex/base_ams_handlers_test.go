// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Source: ts/input/tex/ams/AmsMethods.ts, AmsItems.ts, and AmsMappings.ts.

package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestBaseAMSHandlerFrozenShapes(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		want string
	}{
		{"flalign", `\begin{flalign*}a&=b&&\end{flalign*}`, `math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)]),mtd([]),mtd([]),mtd([])))])`},
		{"genfrac", `\genfrac{(}{)}{1pt}{0}{a}{b}`, `math([mstyle([mrow(TeXAtom([mo(()]),mfrac(mi(a),mi(b)),TeXAtom([mo())]))])])`},
		{"declared-operator", `\DeclareMathOperator{\Foo}{Foo}\Foo(x)`, `math([mi(Foo),mo((),mi(x),mo())])`},
		{"operator-name", `\operatorname*{arg\,max}_{x}`, `math([munder(TeXAtom([mi(arg),mstyle([mspace()]),mi(max)]),TeXAtom([mi(x)]))])`},
		{"boxed", `\boxed{x}`, `math([menclose([TeXAtom([mstyle([TeXAtom([mi(x)])])])])])`},
		{"multi-integral", `\idotsint_A f`, `math([mo(∫),mo(⋯),msub(mo(∫),mi(A)),mi(f)])`},
		{"sideset", `\sideset{_a^b}{_c^d}\sum`, `math([TeXAtom([mmultiscripts(mo(∑),mi(c),mi(d),mprescripts(),mi(a),mi(b))])])`},
		{"x-arrow", `\xrightarrow[below]{above}`, `math([munderover(mstyle([mo(→)]),mpadded([mi(b),mi(e),mi(l),mi(o),mi(w),mspace()]),mpadded([mi(a),mi(b),mi(o),mi(v),mi(e),mspace()]))])`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := mmlString(root); got != test.want {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.want)
			}
		})
	}
}

func TestBaseAMSXArrowAttributeOrder(t *testing.T) {
	p := &parser{source: `[below]{above}`, state: newParseState(), display: true}
	nodes, handled, err := p.baseAMSCommand("xrightarrow")
	if err != nil || !handled {
		t.Fatalf("xrightarrow handled=%v err=%v", handled, err)
	}
	arrow := nodes[0]
	if arrow.Kind != "munderover" || len(arrow.Children) != 3 {
		t.Fatalf("xrightarrow root = %#v", arrow)
	}
	under, over := arrow.Children[1], arrow.Children[2]
	if got, want := under.Attributes.ExplicitNames(), []string{"width", "lspace", "voffset", "depth"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("under attributes = %v, want %v", got, want)
	}
	if got, want := over.Attributes.ExplicitNames(), []string{"width", "lspace", "voffset", "height"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("over attributes = %v, want %v", got, want)
	}
	mo := arrow.Children[0].Children[0].Children[0]
	if mo.Kind != "mo" || mo.TeXClass != mml.TeXClassRel {
		t.Fatalf("arrow token = %#v", mo)
	}
}

func TestBaseAMSHandlerErrors(t *testing.T) {
	p := &parser{source: `{(}{)}{1pt}{9}{a}{b}`, state: newParseState(), display: true}
	_, handled, err := p.baseAMSCommand("genfrac")
	if !handled {
		t.Fatal("genfrac was not handled")
	}
	texErr, ok := err.(*Error)
	if !ok || texErr.ID != "BadMathStyleFor" {
		t.Fatalf("error = %#v, want BadMathStyleFor", err)
	}
}
