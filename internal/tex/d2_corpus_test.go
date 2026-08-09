// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// These are the 17 parser trees produced by D2's frozen MathJax 3.2.2 bundle
// for TestFrozenMathJaxParity. They were extracted at STATE.COMPILED (before
// output-jax layout) using MmlNode.toString().
var d2FrozenMML = []struct {
	name, tex, mml string
}{
	{"basic", `a + b = c`, `math([mi(a),mo(+),mi(b),mo(=),mi(c)])`},
	{"fraction", `\frac{1}{2}`, `math([mfrac(mn(1),mn(2))])`},
	{"multiline", "a + b\n= c\n", `math([mi(a),mo(+),mi(b),mo(=),mi(c)])`},
	{"scripts", `x_i^2+\sum_{n=0}^{\infty}n^{-2}`, `math([msubsup(mi(x),mi(i),mn(2)),mo(+),munderover(mo(∑),TeXAtom([mi(n),mo(=),mn(0)]),TeXAtom([mi(∞)])),msup(mi(n),TeXAtom([mo(−),mn(2)]))])`},
	{"delimiters", `\left\langle \frac{x}{y}\middle|z\right\rangle`, `math([mrow(mo(⟨),mfrac(mi(x),mi(y)),TeXAtom([]),mo(|),TeXAtom([]),mi(z),mo(⟩))])`},
	{"ams_matrix", `\begin{pmatrix}a&b\\c&d\end{pmatrix}`, `math([mrow(mo((),mtable(mtr(mtd([mi(a)]),mtd([mi(b)])),mtr(mtd([mi(c)]),mtd([mi(d)]))),mo()))])`},
	{"ams_align", `\begin{aligned}a&=b+c\\d&=e\end{aligned}`, `math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b),mo(+),mi(c)])),mtr(mtd([mi(d)]),mtd([mi(),mo(=),mi(e)])))])`},
	{"mathtools", `\min_{\mathclap{\substack{x\in\mathbb{R}^n\\x\geq0\\Ax\leq b}}}c^Tx`, `math([munder(mo(min),TeXAtom([TeXAtom([mstyle([mpadded([mtable(mtr(mtd([mi(x),mo(∈),msup(TeXAtom([mi(R)]),mi(n))])),mtr(mtd([mi(x),mo(≥),mn(0)])),mtr(mtd([mi(A),mi(x),mo(≤),mi(b)])))])])])])),msup(mi(c),mi(T)),mi(x)])`},
	{"amscd", `\begin{CD}A@>f>>B\\@VgVV@VVhV\\C@>>k>D\end{CD}`, `math([mtable(mtr(mtd([mi(A),mpadded([])]),mtd([mover(mo(→),mpadded([mi(f)]))]),mtd([mi(B)])),mtr(mtd([mrow(mstyle([TeXAtom([mpadded([mi(g)])])]),mo(↓))]),mtd([]),mtd([mrow(mo(↓),mstyle([TeXAtom([mpadded([mi(h)])])]))]),mtd([])),mtr(mtd([mi(C),mpadded([])]),mtd([munderover(mo(→),mpadded([mi(k)]),mpadded([mspace()]))]),mtd([mi(D)])))])`},
	{"braket", `\bra{\psi}A\ket{\phi}+\braket{x|y}`, `math([mrow(mo(⟨),TeXAtom([mi(ψ)]),mo(|)),TeXAtom([]),mi(A),mrow(mo(|),TeXAtom([mi(ϕ)]),mo(⟩)),mo(+),mrow(mo(⟨),TeXAtom([mi(x),mo(|),mi(y)]),TeXAtom([]),mo(|),TeXAtom([]),TeXAtom([mi(x),mo(|),mi(y)]),mo(⟩))])`},
	{"cancel", `\cancel{x}+\bcancel{y}+\xcancel{z}+\cancelto{0}{w}`, `math([menclose([mi(x)]),mo(+),menclose([mi(y)]),mo(+),menclose([mi(z)]),mo(+),msup(menclose([mi(w)]),mpadded([mn(0)]))])`},
	{"cases", `f(x)=\begin{cases}x^2&x\geq0\\-x&x<0\end{cases}`, `math([mi(f),mo((),mi(x),mo()),mo(=),mrow(mo({),mtable(mtr(mtd([msup(mi(x),mn(2))]),mtd([mi(x),mo(≥),mn(0)])),mtr(mtd([mo(−),mi(x)]),mtd([mi(x),mo(<),mn(0)]))),mo())])`},
	{"color", `\color{red}{x}+\colorbox{yellow}{y}`, `math([mstyle([TeXAtom([mi(x)]),mo(+),mpadded([mtext(y)])])])`},
	{"gensymb", `90\degree+20\celsius+5\micro\ohm`, `math([mn(90),mi(°),mo(+),mn(20),mi(℃),mo(+),mn(5),mi(µ),mi(Ω)])`},
	{"mhchem", `\ce{2H2 + O2 -> 2H2O}`, `math([TeXAtom([mn(2),mstyle([mspace()]),TeXAtom([mi(H)]),msub(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([TeXAtom([mpadded([mn(2)])])])),TeXAtom([]),mo(+),TeXAtom([]),TeXAtom([mi(O)]),msub(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([TeXAtom([mpadded([mn(2)])])])),TeXAtom([]),TeXAtom([mo(⟶)]),TeXAtom([]),mn(2),mstyle([mspace()]),TeXAtom([mi(H)]),msub(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([TeXAtom([mpadded([mn(2)])])])),TeXAtom([mi(O)])])])`},
	{"physics", `\dv{x}f(x)+\qty(\frac{a}{b})`, `math([mfrac(TeXAtom([mi(d)]),mrow(TeXAtom([mi(d)]),mi(x))),mi(f),mo((),mi(x),mo()),mo(+),mrow(mo((),mfrac(mi(a),mi(b)),mo()))])`},
	{"invalid_merror", `\frac{1}{`, `math([merror([mtext(Missing close brace)])])`},
}

func TestD2FrozenMMLCorpus(t *testing.T) {
	compiler := NewCompiler()
	for _, test := range d2FrozenMML {
		t.Run(test.name, func(t *testing.T) {
			root, err := compiler.Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := mmlString(root); got != test.mml {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.mml)
			}
		})
	}
}

func mmlString(node *mml.Node) string {
	if node.Kind == "text" {
		return node.Text
	}
	children := make([]string, len(node.Children))
	for i, child := range node.Children {
		children[i] = mmlString(child)
	}
	if node.Kind == "mrow" && node.Flags.Inferred {
		return "[" + strings.Join(children, ",") + "]"
	}
	return node.Kind + "(" + strings.Join(children, ",") + ")"
}
