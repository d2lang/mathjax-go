// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/oracle"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func token(kind, text string, class mml.TeXClass) *mml.Node {
	node := mml.NewNode(kind, nil, nil, mml.NewText(text))
	node.Flags.Token = true
	node.TeXClass = class
	if kind == "mo" {
		node.Flags.Embellished = true
	}
	return node
}

func simpleFormula() *mml.Node {
	row := mml.NewNode("mrow", nil, nil,
		token("mi", "a", mml.TeXClassOrd),
		token("mo", "+", mml.TeXClassBin),
		token("mi", "b", mml.TeXClassOrd),
		token("mo", "=", mml.TeXClassRel),
		token("mi", "c", mml.TeXClassOrd),
	)
	row.Flags.Inferred = true
	root := mml.NewNode("math", nil, nil, row)
	root.Attributes.Set("display", "block")
	return root
}

func fractionFormula() *mml.Node {
	numerator := token("mi", "a", mml.TeXClassOrd)
	denominator := token("mi", "b", mml.TeXClassOrd)
	fraction := mml.NewNode("mfrac", nil, nil, numerator, denominator)
	fraction.TeXClass = mml.TeXClassInner
	row := mml.NewNode("mrow", nil, nil, fraction)
	row.Flags.Inferred = true
	root := mml.NewNode("math", nil, nil, row)
	root.Attributes.Set("display", "block")
	return root
}

func scriptFormula(kind string) *mml.Node {
	base := token("mi", "x", mml.TeXClassOrd)
	children := []*mml.Node{base}
	switch kind {
	case "msub":
		children = append(children, token("mn", "1", mml.TeXClassOrd))
	case "msup":
		children = append(children, token("mn", "2", mml.TeXClassOrd))
	case "msubsup":
		children = append(children,
			token("mn", "1", mml.TeXClassOrd),
			token("mn", "2", mml.TeXClassOrd),
		)
	}
	script := mml.NewNode(kind, nil, nil, children...)
	script.TeXClass = mml.TeXClassOrd
	row := mml.NewNode("mrow", nil, nil, script)
	row.Flags.Inferred = true
	root := mml.NewNode("math", nil, nil, row)
	root.Attributes.Set("display", "block")
	return root
}

func rootFormula(kind string) *mml.Node {
	base := token("mi", "x", mml.TeXClassOrd)
	var radical *mml.Node
	if kind == "msqrt" {
		row := mml.NewNode("mrow", nil, nil, base)
		row.Flags.Inferred = true
		radical = mml.NewNode("msqrt", nil, nil, row)
	} else {
		radical = mml.NewNode("mroot", nil, nil, base, token("mn", "3", mml.TeXClassOrd))
	}
	radical.TeXClass = mml.TeXClassOrd
	row := mml.NewNode("mrow", nil, nil, radical)
	row.Flags.Inferred = true
	root := mml.NewNode("math", nil, nil, row)
	root.Attributes.Set("display", "block")
	return root
}

func TestSimpleFormulaFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: "a+b=c", Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(simpleFormula(), options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("SVG differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestFractionFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: `\frac{a}{b}`, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(fractionFormula(), options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("SVG differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestScriptsFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	cases := []oracle.Case{
		{TeX: `x_1`, Options: options},
		{TeX: `x^2`, Options: options},
		{TeX: `x_1^2`, Options: options},
	}
	want, err := runner.RenderBatch(withTimeout(t), cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range []string{"msub", "msup", "msubsup"} {
		got, err := svg.NewTypesetter().Typeset(scriptFormula(kind), options)
		if err != nil {
			t.Fatal(err)
		}
		if got != want[i].SVG {
			t.Errorf("%s differs from frozen MathJax 3.2.2:\n got %s\nwant %s", kind, got, want[i].SVG)
		}
	}
}

func TestRootsFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	cases := []oracle.Case{
		{TeX: `\sqrt{x}`, Options: options},
		{TeX: `\sqrt[3]{x}`, Options: options},
	}
	want, err := runner.RenderBatch(withTimeout(t), cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range []string{"msqrt", "mroot"} {
		got, err := svg.NewTypesetter().Typeset(rootFormula(kind), options)
		if err != nil {
			t.Fatal(err)
		}
		if got != want[i].SVG {
			t.Errorf("%s differs from frozen MathJax 3.2.2:\n got %s\nwant %s", kind, got, want[i].SVG)
		}
	}
}

func TestCompiledScriptsFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	const source = `x_i^2+\sum_{n=0}^{\infty}n^{-2}`
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: source, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("compiled scripts differ from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestCompiledMerrorFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	const source = `\frac{1}{`
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: source, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("compiled merror differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestCompiledDelimitersFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	const source = `\left\langle \frac{x}{y}\middle|z\right\rangle`
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: source, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("compiled delimiters differ from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestCompiledPaddedAndPhantomFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	cases := []oracle.Case{
		{TeX: `\mathclap{x}`, Options: options},
		{TeX: `\phantom{x}`, Options: options},
	}
	want, err := runner.RenderBatch(withTimeout(t), cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, source := range []string{`\mathclap{x}`, `\phantom{x}`} {
		root, err := tex.NewCompiler().Compile(source, true)
		if err != nil {
			t.Fatal(err)
		}
		got, err := svg.NewTypesetter().Typeset(root, options)
		if err != nil {
			t.Fatal(err)
		}
		if got != want[i].SVG {
			t.Errorf("%s differs from frozen MathJax 3.2.2:\n got %s\nwant %s", source, got, want[i].SVG)
		}
	}
}

func TestCompiledTeXClassTraversalFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	options := pipeline.DefaultOptions()
	formulas := []struct {
		name string
		tex  string
	}{
		{"fraction_followed_by_identifier", `\dv{x}f(x)+\qty(\frac{a}{b})`},
		{"left_right_row_followed_by_identifier", `\bra{\psi}A\ket{\phi}+\braket{x|y}`},
	}
	cases := make([]oracle.Case, len(formulas))
	for i, formula := range formulas {
		cases[i] = oracle.Case{TeX: formula.tex, Options: options}
	}
	want, err := runner.RenderBatch(withTimeout(t), cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, formula := range formulas {
		t.Run(formula.name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(formula.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != want[i].SVG {
				t.Fatalf("compiled TeX-class traversal differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[i].SVG)
			}
		})
	}
}

func TestCompiledApplyFunctionFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	const source = `\sin x`
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: source, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("compiled ApplyFunction spacing differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func TestCompiledHugeFixedDelimiterScriptFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	const source = `\Huge{\frac{\alpha g^2}{\omega^5} e^{[ -0.74\bigl\{\frac{\omega U_\omega 19.5}{g}\bigr\}^{\!-4}\,]}}`
	options := pipeline.DefaultOptions()
	want, err := runner.RenderBatch(withTimeout(t), []oracle.Case{{TeX: source, Options: options}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != want[0].SVG {
		t.Fatalf("compiled Huge fixed-delimiter script differs from frozen MathJax 3.2.2:\n got %s\nwant %s", got, want[0].SVG)
	}
}

func withTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}
