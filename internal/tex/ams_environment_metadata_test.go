// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Sources: ts/input/tex/ams/AmsMethods.ts and AmsItems.ts.

package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestAMSFullWidthEnvironmentMetadata(t *testing.T) {
	tests := []struct {
		name       string
		tex        string
		attributes []string
		values     map[string]any
		cellAlign  []string
	}{
		{
			name: "xalignat",
			tex:  `\begin{xalignat*}{2}a&=b&c&=d\end{xalignat*}`,
			attributes: []string{
				"width", "displaystyle", "columnalign", "columnspacing",
				"columnwidth", "rowspacing", "minlabelspacing",
				"data-width-includes-label",
			},
			values: map[string]any{
				"width": "100%", "displaystyle": true,
				"columnalign": "center right left center right left center", "columnspacing": "0em",
				"columnwidth": "fit auto auto fit auto auto fit", "rowspacing": "3pt",
				"minlabelspacing": "0", "data-width-includes-label": true,
			},
		},
		{
			name: "multline",
			tex:  `\begin{multline*}a+b\\c+d\end{multline*}`,
			attributes: []string{
				"displaystyle", "rowspacing", "columnspacing", "width",
				"framespacing", "frame", "data-width-includes-label",
			},
			values: map[string]any{
				"displaystyle": true, "rowspacing": ".5em",
				"columnspacing": "100%", "width": "100%",
				"framespacing": "1em 0", "frame": "",
				"data-width-includes-label": true,
			},
			cellAlign: []string{"left", "right"},
		},
		{
			name: "shove-left",
			tex:  `\begin{multline*}a+b\\\shoveleft{c+d}\end{multline*}`,
			attributes: []string{
				"displaystyle", "rowspacing", "columnspacing", "width",
				"framespacing", "frame", "data-width-includes-label",
			},
			values: map[string]any{
				"displaystyle": true, "rowspacing": ".5em",
				"columnspacing": "100%", "width": "100%",
				"framespacing": "1em 0", "frame": "",
				"data-width-includes-label": true,
			},
			cellAlign: []string{"left", "left"},
		},
		{
			name: "flalign",
			tex:  `\begin{flalign*}a&=b&&\end{flalign*}`,
			attributes: []string{
				"width", "displaystyle", "columnalign", "columnspacing",
				"columnwidth", "rowspacing", "data-width-includes-label",
			},
			values: map[string]any{
				"width": "100%", "displaystyle": true,
				"columnalign":   "right left center right left",
				"columnspacing": "0em",
				"columnwidth":   "auto auto fit auto auto",
				"rowspacing":    "3pt", "data-width-includes-label": true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.tex, true)
			if err != nil {
				t.Fatal(err)
			}
			table := onlyAMSTable(t, root)
			if got := table.Attributes.ExplicitNames(); !reflect.DeepEqual(got, test.attributes) {
				t.Fatalf("explicit attributes = %v, want %v", got, test.attributes)
			}
			for name, want := range test.values {
				if got, ok := table.Attributes.GetExplicit(name); !ok || got != want {
					t.Fatalf("%s = %#v, %v; want %#v", name, got, ok, want)
				}
			}
			for row, want := range test.cellAlign {
				cell := table.Children[row].Children[0]
				if got, ok := cell.Attributes.GetExplicit("columnalign"); !ok || got != want {
					t.Fatalf("row %d columnalign = %#v, %v; want %q", row, got, ok, want)
				}
			}
		})
	}
}

func onlyAMSTable(t *testing.T, root *mml.Node) *mml.Node {
	t.Helper()
	tables := root.Find("mtable")
	if len(tables) != 1 {
		t.Fatalf("mtable count = %d, want 1", len(tables))
	}
	return tables[0]
}
