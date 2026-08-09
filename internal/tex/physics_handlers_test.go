// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Source: ts/input/tex/physics/PhysicsMethods.ts and PhysicsMappings.ts.

package tex

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestPhysicsHandlerFrozenShapes(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		want string
	}{
		{"eval", `\eval{x^2}_{0}^{1}`, `math([msubsup(mrow(mo(),msup(mi(x),mn(2)),TeXAtom([mpadded([mphantom([mo(∫)])])]),mo(|)),TeXAtom([mn(0)]),TeXAtom([mn(1)]))])`},
		{"expression", `\sin[2](x)`, `math([msup(mi(sin),mn(2)),mo(⁡),mrow(mo((),mi(x),mo()))])`},
		{"ket-bra", `\outerproduct{x}{y}`, `math([mrow(mo(|),TeXAtom([mi(x)]),TeXAtom([]),mo(⟩),TeXAtom([]),mstyle([mspace()]),TeXAtom([]),mo(⟨),TeXAtom([]),TeXAtom([mi(y)]),mo(|))])`},
		{"matrix-element", `\mel{a}{A}{b}`, `math([mrow(mo(⟨),TeXAtom([mi(a)]),mo(|)),TeXAtom([mi(A)]),mrow(mo(|),TeXAtom([mi(b)]),mo(⟩))])`},
		{"identity-matrix", `\begin{pmatrix}\imat{2}\end{pmatrix}`, `math([mrow(mo((),mtable(mtr(mtd([mn(1)]),mtd([mn(0)])),mtr(mtd([mn(0)]),mtd([mn(1)]))),mo()))])`},
		{"x-matrix", `\begin{pmatrix}\xmat{x}{2}{2}\end{pmatrix}`, `math([mrow(mo((),mtable(mtr(mtd([mi(x)]),mtd([mi(x)])),mtr(mtd([mi(x)]),mtd([mi(x)]))),mo()))])`},
		{"pauli-matrix", `\begin{pmatrix}\pmat{1}\end{pmatrix}`, `math([mrow(mo((),mtable(mtr(mtd([mn(0)]),mtd([mn(1)])),mtr(mtd([mn(1)]),mtd([mn(0)]))),mo()))])`},
		{"diagonal-matrix", `\begin{pmatrix}\dmat{a,b}\end{pmatrix}`, `math([mrow(mo((),mtable(mtr(mtd([mtable(mtr(mtd([mi(a)])))])),mtr(mtd([]),mtd([mtable(mtr(mtd([mi(b)])))]))),mo()))])`},
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

func TestPhysicsVectorOperatorVariants(t *testing.T) {
	tests := []struct {
		tex     string
		glyph   string
		variant any
		set     bool
	}{
		{`\divergence(A)`, "⋅", "bold", true},
		{`\curl(A)`, "×", nil, false},
	}
	for _, test := range tests {
		root, err := NewCompiler().Compile(test.tex, true)
		if err != nil {
			t.Fatal(err)
		}
		var operator *mml.Node
		for _, candidate := range root.Find("mo") {
			if textContent(candidate) == test.glyph {
				operator = candidate
				break
			}
		}
		if operator == nil {
			t.Fatalf("%s: missing %q operator", test.tex, test.glyph)
		}
		got, ok := operator.Attributes.GetExplicit("mathvariant")
		if ok != test.set || got != test.variant {
			t.Fatalf("%s: mathvariant = %#v, present %v; want %#v, present %v", test.tex, got, ok, test.variant, test.set)
		}
	}
}

func TestPhysicsExpressionUsesFixedFenceRow(t *testing.T) {
	root, err := NewCompiler().Compile(`\sin[2](x)`, true)
	if err != nil {
		t.Fatal(err)
	}
	rows := root.Find("mrow")
	var fence *mml.Node
	for _, candidate := range rows {
		if !candidate.Flags.Inferred && textContent(candidate) == "(x)" {
			fence = candidate
			break
		}
	}
	if fence == nil {
		t.Fatal("missing fixed-fence argument row")
	}
	if _, ok := fence.Property("open"); ok {
		t.Fatal("fixed-fence row unexpectedly has an open property")
	}
	if _, ok := fence.Property("close"); ok {
		t.Fatal("fixed-fence row unexpectedly has a close property")
	}
	if fence.TeXClass != mml.TeXClassNone {
		t.Fatalf("fixed-fence row TeX class = %d, want unset", fence.TeXClass)
	}
}

func TestPhysicsMatrixGeneratorSourceShapes(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		command string
		want    string
	}{
		{"indexed-x", `*{x}{2}{2}`, "xmat", `x_{{1}{1}} & x_{{1}{2}}\\ x_{{2}{1}} & x_{{2}{2}}`},
		{"zero", `{2}{3}`, "zmat", `0 & 0 & 0\\ 0 & 0 & 0`},
		{"anti-diagonal", `{a,b,c}`, "admat", `&&\mqty{a}\\ &\mqty{b}\\ \mqty{c}`},
		{"pauli-y", `{y}`, "pmat", ` 0 & -i\\ i & 0`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.source, state: newParseState(), display: true}
			got, err := p.physicsMatrixExpansion(test.command)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("expansion = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPhysicsHandlerErrors(t *testing.T) {
	tests := []struct {
		name    string
		command string
		source  string
		wantID  string
	}{
		{"identity-not-number", "imat", `{x}`, "InvalidNumber"},
		{"x-matrix-noncanonical-size", "xmat", `{x}{02}{2}`, "InvalidNumber"},
		{"eval-missing-argument", "eval", `x`, "MissingArgFor"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.source, state: newParseState(), display: true}
			_, handled, err := p.physicsCommand(test.command)
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

func TestPhysicsMatrixBodyRewriteLeavesOtherCommands(t *testing.T) {
	p := &parser{state: newParseState(), display: true}
	got, err := p.physicsRewriteMatrixBody(`a&\frac{1}{2}\\\imat{2}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, `a&\frac{1}{2}\\`) || !strings.HasSuffix(got, `1 & 0\\ 0 & 1`) {
		t.Fatalf("rewritten body = %q", got)
	}
}
