// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"errors"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestAMSTagsFrozenSourceShapes(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		mml  string
	}{
		{"equation-tag", `\begin{equation}a=b\tag{T}\end{equation}`, `math([mtable(mlabeledtr(mtd([mtext((T))]),mtd([mi(a),mo(=),mi(b)])))])`},
		{"starred-tag", `a\tag*{T}`, `math([mtable(mlabeledtr(mtd([mtext(T)]),mtd([mi(a)])))])`},
		{"equation-reference-placeholder", `\begin{equation}a=b\label{eq:a}\end{equation}\eqref{eq:a}`, `math([mi(a),mo(=),mi(b),mrow(mtext((???)))])`},
		{"notag-align", `\begin{align}a&=b\notag\end{align}`, `math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])`},
		{"top-level-notag", `a\notag`, `math([mtable(mlabeledtr(mtd([]),mtd([mi(a)])))])`},
		{"mathtools-tag-form", `\newtagform{brackets}{[}{]}\usetagform{brackets}\begin{equation}a=b\tag{T}\end{equation}`, `math([mtable(mlabeledtr(mtd([mtext([T])]),mtd([mi(a),mo(=),mi(b)])))])`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := mmlString(root); got != test.mml {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.mml)
			}
		})
	}
}

func TestAMSTagErrors(t *testing.T) {
	tests := []struct {
		name    string
		tex     string
		wantID  string
		message string
	}{
		{"tag-disallowed", `{T}`, "CommandNotAllowedInEnv", `\tag not allowed in aligned environment`},
		{"duplicate-tag", `{B}`, "MultipleCommand", `Multiple \tag`},
		{"duplicate-label", `{same}`, "MultipleLabel", `Label 'same' multiply defined`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.tex, state: newParseState(), display: true}
			state := p.amsTags()
			var err error
			switch test.name {
			case "tag-disallowed":
				state.start("aligned", false, false)
				err = p.amsHandleTag("tag")
			case "duplicate-tag":
				state.start("equation", true, true)
				state.setTag(p, "A", false)
				err = p.amsHandleTag("tag")
			case "duplicate-label":
				state.labels["same"] = amsLabel{}
				err = p.amsHandleLabel("label")
			}
			var texErr *Error
			if !errors.As(err, &texErr) {
				t.Fatalf("error = %v, want *Error", err)
			}
			if texErr.ID != test.wantID || texErr.Message != test.message {
				t.Fatalf("error = (%q, %q), want (%q, %q)", texErr.ID, texErr.Message, test.wantID, test.message)
			}
		})
	}
}

func TestAMSEquationSplitFrozenShape(t *testing.T) {
	const source = "\\begin{equation} \\label{eq1}\n\\begin{split}\nA & = \\frac{\\\\pi r^2}{2} \\\\\n & = \\frac{1}{2} \\pi r^2\n\\end{split}\n\\end{equation}"
	const want = `math([mtable(mtr(mtd([mi(A)]),mtd([mi(),mo(=),mfrac(mrow(mspace(),mi(p),mi(i),msup(mi(r),mn(2))),mn(2))])),mtr(mtd([]),mtd([mi(),mo(=),mfrac(mn(1),mn(2)),mi(π),msup(mi(r),mn(2))])))])`
	root, err := NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := mmlString(root); got != want {
		t.Fatalf("MML mismatch:\n got %s\nwant %s", got, want)
	}
	tables := root.Find("mtable")
	if len(tables) != 1 {
		t.Fatalf("mtable count = %d, want 1", len(tables))
	}
	if got, want := tables[0].Attributes.ExplicitNames(), []string{"displaystyle", "columnalign", "columnspacing", "rowspacing"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("mtable explicit attributes = %v, want %v", got, want)
	}
	if value, ok := tables[0].Property("useHeight"); !ok || value != true {
		t.Fatalf("mtable useHeight = %#v, %v", value, ok)
	}
}

func TestAMSTagReferenceAttributes(t *testing.T) {
	p := &parser{source: `{x y}`, state: newParseState(), display: true}
	p.amsTags().labels["x y"] = amsLabel{tag: "7", id: "mjx-eqn:x y"}
	nodes, err := p.amsHandleReference("ref", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Kind != "mrow" {
		t.Fatalf("reference nodes = %#v", nodes)
	}
	if href, _ := nodes[0].Attributes.GetExplicit("href"); href != "#mjx-eqn%3Ax%20y" {
		t.Fatalf("href = %#v", href)
	}
	if class, _ := nodes[0].Attributes.GetExplicit("class"); class != "MathJax_ref" {
		t.Fatalf("class = %#v", class)
	}
	if nodes[0].TeXClass != mml.TeXClassNone {
		t.Fatalf("reference tex class = %d, want unset", nodes[0].TeXClass)
	}
}
