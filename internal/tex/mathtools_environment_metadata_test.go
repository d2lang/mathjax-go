// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Sources: ts/input/tex/mathtools/MathtoolsMethods.ts,
// MathtoolsItems.ts, and MathtoolsUtil.ts.

package tex

import (
	"strings"
	"testing"
)

func TestMathtoolsGatherRestoresEnclosingTagState(t *testing.T) {
	for _, environment := range []string{"gather", "gather*"} {
		for _, test := range []struct {
			name, body, errorText string
		}{
			{"success", `x=y\tag{T}\\\vdotswithin{=}`, ""},
			{"duplicate-tag", `x\tag{T}\tag{U}\\\vdotswithin{=}`, `Multiple \tag`},
			{"special-row-duplicate-tag", `\vdotswithin{=}\tag{T}\tag{U}`, `Multiple \tag`},
			{"nested-equation", `\begin{equation}x\end{equation}\\\vdotswithin{=}`, "Erroneous nesting"},
		} {
			t.Run(environment+"/"+test.name, func(t *testing.T) {
				p := &parser{source: test.body + `\end{` + environment + `}`, state: newParseState(), display: true}
				tags := p.amsTags()
				outer := tags.current
				outer.tag = stringPointer("outer")
				outer.tagFormat = "outer"
				_, err := p.amsAlignment(environment)
				if test.errorText == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), test.errorText) {
					t.Fatalf("error = %v, want %q", err, test.errorText)
				}
				if tags.current != outer || *outer.tag != "outer" || outer.tagFormat != "outer" || len(tags.stack) != 0 {
					t.Fatal("gather did not restore the enclosing tag state")
				}
				if len(tags.history) != 1 || tags.history[0].environment != environment {
					t.Fatal("gather tag environment did not end exactly once")
				}
			})
		}
	}
}

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

func TestMathtoolsGatherPreservesInteriorBlankRow(t *testing.T) {
	root, err := NewCompiler().Compile(`\begin{gather}x\\ \\ \vdotswithin{=}\end{gather}`, true)
	if err != nil {
		t.Fatal(err)
	}
	if rows := len(onlyAMSTable(t, root).Children); rows != 3 {
		t.Fatalf("gather has %d rows, want the authored interior blank row", rows)
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
