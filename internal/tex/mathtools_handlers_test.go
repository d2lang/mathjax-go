// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts,
// MathtoolsItems.ts, and MathtoolsUtil.ts.

package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestMathtoolsHandlerSourceShapes(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		want string
	}{
		{
			"paired-delimiter-xpp",
			`\DeclarePairedDelimiterXPP{\pairpp}[2]{P}{(}{)}{Q}{#1,#2}\pairpp{a}{b}`,
			`math([mi(P),mo((),mi(a),mo(,),mi(b),mo()),mi(Q)])`,
		},
		{
			"overbracket-options",
			`\overbracket[.2em][.3em]{x+y}`,
			`math([TeXAtom([mover(mrow(mi(x),mo(+),mi(y)),mpadded([mphantom([mi(x),mo(+),mi(y)])]))])])`,
		},
		{
			"multlined",
			`\begin{multlined}[c]a+b\\c+d\end{multlined}`,
			`math([mtable(mtr(mtd([mi(a),mo(+),mi(b),mspace()])),mtr(mtd([mspace(),mi(c),mo(+),mi(d)])))])`,
		},
		{
			"alignment-methods",
			`\begin{align*}a&=b\\\ArrowBetweenLines\\\vdotswithin{=}\\c&=d\end{align*}`,
			`math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])),mtr(mtd([mo(⇕),mstyle([mspace()])])),mtr(mtd([])),mtr(mtd([mpadded([mpadded([mo(⋮)]),mphantom([mi(),mo(=),mi()])])])),mtr(mtd([mi(c)]),mtd([mi(),mo(=),mi(d)])))])`,
		},
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

func TestMathtoolsHandlerOrderedAttributes(t *testing.T) {
	p := &parser{source: "", state: newParseState(), display: true}
	colon := p.mathtoolsCenterColon(true, true, true)
	if got, want := colon.Attributes.ExplicitNames(), []string{"voffset", "height", "depth", "width", "lspace"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("center-colon attribute order = %v, want %v", got, want)
	}

	arrow := mathtoolsNArrow("nuparrow")
	if arrow.Kind != "TeXAtom" || arrow.TeXClass != mml.TeXClassRel || len(arrow.Children) != 1 {
		t.Fatalf("nuparrow root = %#v", arrow)
	}
	content := arrow.Children[0]
	if content.Kind != "mrow" || len(content.Children) != 2 {
		t.Fatalf("nuparrow inferred content = %#v", content)
	}
	outer := content.Children[1]
	if got, want := outer.Attributes.ExplicitNames(), []string{"width", "lspace"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nuparrow outer attributes = %v, want %v", got, want)
	}
}

func TestMathtoolsBracketUsesStackedOperatorAtom(t *testing.T) {
	root, err := NewCompiler().Compile(`\overbracket[.2em][.3em]{x+y}`, true)
	if err != nil {
		t.Fatal(err)
	}
	atoms := root.Find("TeXAtom")
	if len(atoms) == 0 {
		t.Fatal("missing bracket TeXAtom")
	}
	atom := atoms[0]
	if atom.TeXClass != mml.TeXClassOp {
		t.Fatalf("bracket TeX class = %d, want OP", atom.TeXClass)
	}
	for _, property := range []string{"movesupsub", "subsupOK"} {
		if value, ok := atom.Property(property); !ok || value != true {
			t.Fatalf("bracket %s = %#v, present %v", property, value, ok)
		}
	}
}

func TestMathtoolsHandlerErrors(t *testing.T) {
	tests := []struct {
		name    string
		command string
		source  string
		wantID  string
	}{
		{"xmathstrut-not-number", "xmathstrut", `{nope}`, "NotANumber"},
		{"invalid-option", "mathtoolsset", `{not-an-option=true}`, "InvalidOption"},
		{"outside-alignment", "Aboxed", `{x}`, "NotInAlignment"},
		{"undefined-tag-form", "usetagform", `{missing}`, "UndefinedTagForm"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.source, state: newParseState(), display: true}
			_, handled, err := p.mathtoolsCommand(test.command)
			if !handled {
				t.Fatal("command was not handled")
			}
			texErr, ok := err.(*Error)
			if !ok || texErr.ID != test.wantID {
				t.Fatalf("error = %#v, want ID %q", err, test.wantID)
			}
		})
	}
}

func TestMathtoolsTagFormState(t *testing.T) {
	state := newParseState()
	p := &parser{source: `{brackets}{[}{]}`, state: state, display: true}
	if _, handled, err := p.mathtoolsCommand("newtagform"); !handled || err != nil {
		t.Fatalf("newtagform handled=%v err=%v", handled, err)
	}
	p = &parser{source: `{brackets}`, state: state, display: true}
	if _, handled, err := p.mathtoolsCommand("usetagform"); !handled || err != nil {
		t.Fatalf("usetagform handled=%v err=%v", handled, err)
	}
	if got, want := p.mathtoolsFormatTag("T"), "[T]"; got != want {
		t.Fatalf("formatted tag = %q, want %q", got, want)
	}
}
