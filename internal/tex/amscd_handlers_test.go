// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Source: ts/input/tex/amscd/AmsCdMethods.ts and AmsCdConfiguration.ts.

package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestAmsCDArrowDimensions(t *testing.T) {
	state := newParseState()
	setting := &parser{source: `{3em}`, state: state, display: true}
	if _, handled, err := setting.amscdCommand("minCDarrowwidth"); !handled || err != nil {
		t.Fatalf("minCDarrowwidth handled=%v err=%v", handled, err)
	}
	cd := &parser{source: `A@>>>B\end{CD}`, state: state, display: true}
	nodes, handled, err := cd.amscdEnvironment("CD")
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("CD was not handled")
	}
	root := node("math", nodes...)
	want := `math([mtable(mtr(mtd([mi(A),mpadded([])]),mtd([mover(mo(→),mpadded([mspace()]))]),mtd([mi(B)])))])`
	if got := mmlString(root); got != want {
		t.Fatalf("MML mismatch:\n got %s\nwant %s", got, want)
	}
	var arrow *mml.Node
	root.Walk(func(current *mml.Node) bool {
		if current.Kind == "mo" && textContent(current) == "→" {
			arrow = current
		}
		return true
	})
	if arrow == nil {
		t.Fatal("horizontal arrow not found")
	}
	if got, want := arrow.Attributes.ExplicitNames(), []string{"minsize", "stretchy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("arrow attribute order = %v, want %v", got, want)
	}
	if value, _ := arrow.Attributes.GetExplicit("minsize"); value != "3em" {
		t.Fatalf("minsize = %v, want 3em", value)
	}
	spaces := root.Find("mspace")
	if len(spaces) != 1 {
		t.Fatalf("mspace count = %d, want 1", len(spaces))
	}
	if value, _ := spaces[0].Attributes.GetExplicit("width"); value != "3em" {
		t.Fatalf("empty-label mspace width = %v, want 3em", value)
	}
}

func TestAmsCDVerticalDefaults(t *testing.T) {
	p := &parser{source: `A@VVV\\B\end{CD}`, state: newParseState(), display: true}
	nodes, handled, err := p.amscdEnvironment("CD")
	if err != nil || !handled {
		t.Fatalf("CD handled=%v err=%v", handled, err)
	}
	var arrow *mml.Node
	nodes[0].Walk(func(current *mml.Node) bool {
		if current.Kind == "mo" && textContent(current) == "↓" {
			arrow = current
		}
		return true
	})
	if arrow == nil {
		t.Fatal("vertical arrow not found")
	}
	if value, _ := arrow.Attributes.GetExplicit("minsize"); value != "1.75em" {
		t.Fatalf("minsize = %v, want 1.75em", value)
	}
	if got, want := arrow.Attributes.ExplicitNames(), []string{"minsize", "stretchy", "symmetric", "lspace", "rspace"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("vertical attribute order = %v, want %v", got, want)
	}
}

func TestAmsCDVerticalTailFillsPendingCell(t *testing.T) {
	root, err := NewCompiler().Compile(`\begin{CD} A\\ @VVV V \end{CD}`, true)
	if err != nil {
		t.Fatal(err)
	}
	rows := root.Find("mtr")
	if len(rows) != 2 {
		t.Fatalf("mtr count = %d, want 2", len(rows))
	}
	want := `mtr(mtd([mo(↓)]),mtd([mi(V)]))`
	if got := mmlString(rows[1]); got != want {
		t.Fatalf("arrow row MML = %s, want %s", got, want)
	}
}
