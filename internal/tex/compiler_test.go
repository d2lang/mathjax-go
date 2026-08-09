// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"sync"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// TestFrozenMMLFormula is the first parser-side frozen oracle: it records the
// MathML tree produced by MathJax 3.2.2 before SVG layout for the same simple
// expression used by internal/oracle and internal/svg.
func TestFrozenMMLFormula(t *testing.T) {
	root, err := NewCompiler().Compile("a+b=c", true)
	if err != nil {
		t.Fatal(err)
	}
	if root.Kind != "math" || len(root.Children) != 1 {
		t.Fatalf("root = %s with %d children", root.Kind, len(root.Children))
	}
	display, _ := root.Attributes.GetExplicit("display")
	if display != "block" {
		t.Fatalf("display = %#v, want block", display)
	}
	row := root.Children[0]
	if row.Kind != "mrow" || !row.Flags.Inferred || len(row.Children) != 5 {
		t.Fatalf("row = %#v", row)
	}
	wantKind := []string{"mi", "mo", "mi", "mo", "mi"}
	wantText := []string{"a", "+", "b", "=", "c"}
	wantClass := []mml.TeXClass{mml.TeXClassOrd, mml.TeXClassBin, mml.TeXClassOrd, mml.TeXClassRel, mml.TeXClassOrd}
	for i, child := range row.Children {
		if child.Kind != wantKind[i] || textContent(child) != wantText[i] || child.TeXClass != wantClass[i] {
			t.Errorf("child %d = %s %q class %d; want %s %q class %d", i, child.Kind, textContent(child), child.TeXClass, wantKind[i], wantText[i], wantClass[i])
		}
	}
}

func TestFractionScriptsAndRoot(t *testing.T) {
	root, err := NewCompiler().Compile(`\frac{x_1^2}{\sqrt[3]{y}}`, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(root.Find("mfrac")); got != 1 {
		t.Fatalf("mfrac count = %d", got)
	}
	if got := len(root.Find("msubsup")); got != 1 {
		t.Fatalf("msubsup count = %d", got)
	}
	if got := len(root.Find("mroot")); got != 1 {
		t.Fatalf("mroot count = %d", got)
	}
}

func TestLoadedPackageExamples(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		kind string
	}{
		{"braket", `\bra{a}\ket{b}`, "mrow"},
		{"cancel", `\cancel{Culture + 5}`, "menclose"},
		{"color", `\textcolor{red}{y}`, "mstyle"},
		{"gensymb", `10.6\,\micro\mathrm{m}`, "mi"},
		{"physics", `\var{F[g(x)]}\dd(\cos\theta)`, "mrow"},
		{"mathtools", `\mathclap{\substack{x\in\mathbb{R}^n\\x\geq0}}`, "mpadded"},
		{"ams", `\begin{pmatrix}a&b\\c&d\end{pmatrix}`, "mtable"},
	}
	compiler := NewCompiler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := compiler.Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			if errors := root.Find("merror"); len(errors) != 0 {
				t.Fatalf("unexpected merror: %q", textContent(errors[0]))
			}
			if got := len(root.Find(test.kind)); got == 0 {
				t.Fatalf("missing %s in tree; text %q", test.kind, textContent(root))
			}
		})
	}
}

func TestTeXErrorsBecomeMError(t *testing.T) {
	root, err := NewCompiler().Compile(`\frac{1}{`, true)
	if err != nil {
		t.Fatalf("Compile returned Go error: %v", err)
	}
	errors := root.Find("merror")
	if len(errors) != 1 {
		t.Fatalf("merror count = %d", len(errors))
	}
	if got := textContent(errors[0]); got != "Missing close brace" {
		t.Fatalf("merror text = %q", got)
	}
	if value, _ := errors[0].Attributes.GetExplicit("data-mjx-error"); value != "Missing close brace" {
		t.Fatalf("data-mjx-error = %#v", value)
	}
}

func TestCompileStateIsFreshAndConcurrent(t *testing.T) {
	// Macro-definition state belongs to the additive newcommand oracle rather
	// than D2's public package set; the freshness/concurrency invariant applies
	// identically to this private compiler mode.
	compiler := &Compiler{augmentedPackages: true}
	defined, err := compiler.Compile(`\newcommand{\foo}{x}\foo`, false)
	if err != nil || len(defined.Find("merror")) != 0 || textContent(defined) != "x" {
		t.Fatalf("definition compile = %q, %v", textContent(defined), err)
	}
	fresh, err := compiler.Compile(`\foo`, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh.Find("merror")) != 1 {
		t.Fatalf("macro leaked across Compile calls")
	}

	const workers = 16
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wait.Done()
			root, err := compiler.Compile(`\definecolor{mine}{RGB}{12,34,56}\textcolor{mine}{x}`, false)
			if err != nil || len(root.Find("merror")) != 0 {
				t.Errorf("concurrent compile: %v, %q", err, textContent(root))
			}
		}()
	}
	wait.Wait()
}

func TestD2ConfiguredPackageBoundary(t *testing.T) {
	tests := []struct {
		name    string
		tex     string
		message string
	}{
		{"newcommand", `\newcommand{\foo}{x}\foo`, `Undefined control sequence \newcommand`},
		{"def", `\def\foo{x}\foo`, `Undefined control sequence \def`},
		{"let", `\let\foo=\alpha\foo`, `Undefined control sequence \let`},
		{"empheq environment", `\begin{empheq}{align}a&=b\end{empheq}`, `Unknown environment 'empheq'`},
		{"empheq command", `\empheqlbrace`, `Undefined control sequence \empheqlbrace`},
		{"cases transitive helper", `\begin{numcases}{f(x)=}x&if x>0\\0&otherwise\end{numcases}`, `Undefined control sequence \empheqlbrace`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			errors := root.Find("merror")
			if len(errors) != 1 || textContent(errors[0]) != test.message {
				t.Fatalf("merror = %q, want %q", textContent(root), test.message)
			}
		})
	}
}
