// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Sources: ts/input/tex/mathtools/MathtoolsMethods.ts,
// MathtoolsItems.ts, and MathtoolsUtil.ts.

package tex

import "testing"

func TestMathtoolsRowSpacingMetadata(t *testing.T) {
	tests := []struct {
		name string
		tex  string
		want string
	}{
		{
			name: "flush-space-above",
			tex:  `\begin{align*}a&=b\\\MTFlushSpaceAbove\vdotswithin{=}\\c&=d\end{align*}`,
			want: "0.1em 0.3em 0.3em",
		},
		{
			name: "flush-space-below",
			tex:  `\begin{align*}a&=b\\\vdotswithin{=}\MTFlushSpaceBelow c&=d\end{align*}`,
			want: "3pt 0.1em 0.3em",
		},
		{
			name: "short-vdots-within",
			tex:  `\begin{align*}a&=b\\\shortvdotswithin{=}\\c&=d\end{align*}`,
			want: "0.1em 0.1em 0.3em 0.3em",
		},
		{
			name: "spreadlines-adds-to-base",
			tex:  `\begin{spreadlines}{1em}\begin{gathered}a\\b\end{gathered}\end{spreadlines}`,
			want: "1.3em",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			table := onlyAMSTable(t, root)
			if got, ok := table.Attributes.GetExplicit("rowspacing"); !ok || got != test.want {
				t.Fatalf("rowspacing = %#v, %v; want %q", got, ok, test.want)
			}
		})
	}
}

func TestMathtoolsMultlinedMetadata(t *testing.T) {
	root, err := NewCompiler().Compile(`\begin{multlined}[c]a+b\\c+d\end{multlined}`, true)
	if err != nil {
		t.Fatal(err)
	}
	table := onlyAMSTable(t, root)
	if got, ok := table.Attributes.GetExplicit("align"); ok {
		t.Fatalf("explicit align = %#v; want source default omitted", got)
	}
	for row, want := range []string{"left", "right"} {
		cell := table.Children[row].Children[0]
		if got, ok := cell.Attributes.GetExplicit("columnalign"); !ok || got != want {
			t.Fatalf("row %d columnalign = %#v, %v; want %q", row, got, ok, want)
		}
	}
}

func TestMathtoolsSmallMatrixMetadata(t *testing.T) {
	root, err := NewCompiler().Compile(`\begin{psmallmatrix*}[r]a&b\\c&d\end{psmallmatrix*}`, true)
	if err != nil {
		t.Fatal(err)
	}
	table := onlyAMSTable(t, root)
	for name, want := range map[string]any{
		"columnalign":   "right",
		"columnspacing": ".333em",
		"rowspacing":    ".2em",
	} {
		if got, ok := table.Attributes.GetExplicit(name); !ok || got != want {
			t.Fatalf("%s = %#v, %v; want %#v", name, got, ok, want)
		}
	}
	if got, ok := table.Attributes.Get("displaystyle"); !ok || got != false {
		t.Fatalf("effective displaystyle = %#v, %v; want false", got, ok)
	}
	for name, want := range map[string]any{"useHeight": false, "scriptlevel": 1} {
		if got, ok := table.Property(name); !ok || got != want {
			t.Fatalf("property %s = %#v, %v; want %#v", name, got, ok, want)
		}
	}
}
