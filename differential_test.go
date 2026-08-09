// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package mathjax

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/d2lang/mathjax-go/internal/oracle"
)

type handlerRenderManifest struct {
	Cases []struct {
		Name string `json:"name"`
		TeX  string `json:"tex"`
	} `json:"cases"`
}

func TestD2CompatibilityCorpus(t *testing.T) {
	if os.Getenv("MATHJAX_GO_ORACLE_FULL") == "" {
		t.Skip("set MATHJAX_GO_ORACLE_FULL=1 to run the full frozen differential corpus")
	}
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}

	formulas := []struct {
		name string
		tex  string
	}{
		{"basic", `a + b = c`},
		{"fraction", `\frac{1}{2}`},
		{"multiline", "a + b\n= c\n"},
		{"scripts", `x_i^2+\sum_{n=0}^{\infty}n^{-2}`},
		{"delimiters", `\left\langle \frac{x}{y}\middle|z\right\rangle`},
		{"ams_matrix", `\begin{pmatrix}a&b\\c&d\end{pmatrix}`},
		{"ams_align", `\begin{aligned}a&=b+c\\d&=e\end{aligned}`},
		{"mathtools", `\min_{\mathclap{\substack{x\in\mathbb{R}^n\\x\geq0\\Ax\leq b}}}c^Tx`},
		{"amscd", `\begin{CD}A@>f>>B\\@VgVV@VVhV\\C@>>k>D\end{CD}`},
		{"d2_amscd_long", `\begin{CD} B @>{\text{very long label}}>> C S^{{\mathcal{W}}_\Lambda}\otimes T @>j>> T\\ @VVV V \end{CD}`},
		{"braket", `\bra{\psi}A\ket{\phi}+\braket{x|y}`},
		{"d2_braket_simple", `\bra{a}\ket{b}`},
		{"d2_grid_outside_labels_braket", `\bra{a}\ket{b}`},
		{"cancel", `\cancel{x}+\bcancel{y}+\xcancel{z}+\cancelto{0}{w}`},
		{"cases", `f(x)=\begin{cases}x^2&x\geq0\\-x&x<0\end{cases}`},
		{"color", `\color{red}{x}+\colorbox{yellow}{y}`},
		{"gensymb", `90\degree+20\celsius+5\micro\ohm`},
		{"mhchem", `\ce{2H2 + O2 -> 2H2O}`},
		{"physics", `\dv{x}f(x)+\qty(\frac{a}{b})`},
		{"invalid_merror", `\frac{1}{`},
		{"d2_huge", `\Huge{\frac{\alpha g^2}{\omega^5} e^{[ -0.74\bigl\{\frac{\omega U_\omega 19.5}{g}\bigr\}^{\!-4}\,]}}`},
		{"d2_emc2", `e = mc^2`},
		{"d2_gibberish_sum", `gibberish\; math:\sum_{i=0}^\infty i^2`},
		{"d2_linear_program", `\min_{ \mathclap{\substack{ x \in \mathbb{R}^n \ x \geq 0 \ Ax \leq b }}} c^T x`},
		{"d2_equation_split", "\\begin{equation} \\label{eq1}\n\\begin{split}\nA & = \\frac{\\\\pi r^2}{2} \\\\\n & = \\frac{1}{2} \\pi r^2\n\\end{split}\n\\end{equation}"},
		{"d2_limit", `\lim_{h \rightarrow 0 } \frac{f(x+h)-f(x)}{h}`},
		{"d2_quadratic", `f(x) = x^2 + 2x + 1`},
		{"d2_physics_plugin", "\\var{F[g(x)]}\n\\dd(\\cos\\theta)"},
		{"d2_displaylines", "\\displaylines{x = a + b \\\\ y = b + c}\n\\sum_{k=1}^{n} h_{k} \\int_{0}^{1} \\bigl(\\partial_{k} f(x_{k-1}+t h_{k} e_{k}) -\\partial_{k} f(a)\\bigr) \\,dt"},
		{"d2_add", `1 + 1`},
	}
	cases := make([]oracle.Case, len(formulas))
	for i, formula := range formulas {
		cases[i] = oracle.Case{TeX: formula.tex, Options: DefaultOptions().pipelineOptions()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	want, err := runner.RenderBatch(ctx, cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, formula := range formulas {
		t.Run(formula.name, func(t *testing.T) {
			if want[i].Error != "" {
				t.Fatalf("frozen MathJax returned error: %s", want[i].Error)
			}
			got, err := Render(formula.tex)
			if err != nil {
				t.Fatal(err)
			}
			if got != want[i].SVG {
				index, gotContext, wantContext := firstSVGDiff(got, want[i].SVG)
				t.Fatalf(
					"SVG differs from frozen MathJax 3.2.2 at byte %d:\n got sha256=%x context=%q\nwant sha256=%x context=%q",
					index, sha256.Sum256([]byte(got)), gotContext,
					sha256.Sum256([]byte(want[i].SVG)), wantContext,
				)
			}
		})
	}
}

func TestRenderOptionsFrozenOracle(t *testing.T) {
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		tex     string
		options Options
	}{
		{"inline", `x_1^2`, Options{Em: 16, Ex: 8, Display: false, FontCache: FontCacheNone}},
		{"custom metrics", `\frac{a}{b}`, Options{Em: 20, Ex: 10, Display: true, FontCache: FontCacheNone}},
		{"inline custom metrics", `\sqrt{x+1}`, Options{Em: 12, Ex: 6, Display: false, FontCache: FontCacheNone}},
	}
	cases := make([]oracle.Case, len(tests))
	for i, test := range tests {
		cases[i] = oracle.Case{TeX: test.tex, Options: test.options.pipelineOptions()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	want, err := runner.RenderBatch(ctx, cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if want[i].Error != "" {
				t.Fatalf("frozen MathJax returned error: %s", want[i].Error)
			}
			got, err := RenderWithOptions(test.tex, test.options)
			if err != nil {
				t.Fatal(err)
			}
			if got != want[i].SVG {
				index, gotContext, wantContext := firstSVGDiff(got, want[i].SVG)
				t.Fatalf("SVG differs at byte %d:\n got %q\nwant %q", index, gotContext, wantContext)
			}
		})
	}
}

func TestSelectedHandlerRenderFrozenOracle(t *testing.T) {
	if os.Getenv("MATHJAX_GO_ORACLE_FULL") == "" {
		t.Skip("set MATHJAX_GO_ORACLE_FULL=1 to run the full frozen differential corpus")
	}
	runner, err := oracle.FromEnvironment()
	if errors.Is(err, oracle.ErrUnavailable) {
		t.Skip("set MATHJAX_GO_ORACLE_DIR to run the frozen differential oracle")
	}
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("internal/tex/testdata/handler_oracle_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest handlerRenderManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Cases) != 107 {
		t.Fatalf("handler render corpus has %d cases, want 107", len(manifest.Cases))
	}
	cases := make([]oracle.Case, len(manifest.Cases))
	for i, test := range manifest.Cases {
		cases[i] = oracle.Case{TeX: test.TeX, Options: DefaultOptions().pipelineOptions()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	want, err := runner.RenderBatch(ctx, cases)
	if err != nil {
		t.Fatal(err)
	}
	for i, test := range manifest.Cases {
		t.Run(test.Name, func(t *testing.T) {
			if want[i].Error != "" {
				t.Fatalf("frozen MathJax returned error: %s", want[i].Error)
			}
			got, err := Render(test.TeX)
			if err != nil {
				t.Fatal(err)
			}
			if got != want[i].SVG {
				index, gotContext, wantContext := firstSVGDiff(got, want[i].SVG)
				t.Fatalf("SVG differs at byte %d:\n got %q\nwant %q", index, gotContext, wantContext)
			}
		})
	}
}

func firstSVGDiff(got, want string) (int, string, string) {
	limit := len(got)
	if len(want) < limit {
		limit = len(want)
	}
	index := 0
	for index < limit && got[index] == want[index] {
		index++
	}
	if index == limit && len(got) == len(want) {
		return -1, "", ""
	}
	start := index - 40
	if start < 0 {
		start = 0
	}
	gotEnd := index + 80
	if gotEnd > len(got) {
		gotEnd = len(got)
	}
	wantEnd := index + 80
	if wantEnd > len(want) {
		wantEnd = len(want)
	}
	return index, got[start:gotEnd], want[start:wantEnd]
}
