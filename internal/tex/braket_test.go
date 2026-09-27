// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
	"reflect"
	"testing"
)

// These controls exercise BraketMethods/BraketItem lifetime rules directly;
// they are source-derived checks, not additional captured primary formulas.
func TestBraketCallerConsumption(t *testing.T) {
	for _, c := range []struct {
		name, source string
		macro        bool
	}{
		{"fraction", `\frac{a}{b}+z`, false},
		{"macro-fraction", `\probe{a}{b}+z`, true},
		{"macro-then-argument", `\probe+z`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			state.macros["probe"] = macroDefinition{body: `\frac{a}{b}`}
			if c.macro {
				state.macros["probe"] = macroDefinition{body: `\frac{#1}{#2}`, arguments: 2}
			}
			p := &parser{source: c.source, state: state}
			nodes, err := p.braket("Braket")
			if err != nil {
				t.Fatal(err)
			}
			if p.source[p.pos:] != "+z" || len(nodes) != 1 || len(nodes[0].Children) != 3 || nodes[0].Children[1].Kind != "mfrac" {
				t.Fatalf("unbraced command did not finish on caller: remaining=%q nodes=%v", p.source[p.pos:], nodes)
			}
		})
	}
}

func TestBraketBarOwnership(t *testing.T) {
	for _, c := range []struct {
		name, source string
		want         []bool
	}{
		{"group-left-and-outer", `\Braket{x{\|}y\left(\|\right)\|z}`, []bool{false, false, true}},
		{"nested-owner", `\Braket{x\Set{a\|b\|c}\|z}`, []bool{true, false, true}},
		{"fixed-limit-does-not-increment", `\set{x\|y\|z}`, []bool{true, true}},
		{"style-barrier", `\Braket{x\displaystyle\|y}`, []bool{false}},
		{"function-barrier", `\Braket{\sin\|x}`, []bool{false}},
		{"not-barrier", `\Braket{\not\|x}`, []bool{false}},
		{"script-barrier", `\Braket{x^\|y}\|`, []bool{false, false}},
		{"after-owner", `\Braket{x\|y}\|`, []bool{true, false}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.source, true)
			if err != nil {
				t.Fatal(err)
			}
			got := []bool{}
			root.Walk(func(n *mml.Node) bool {
				if n.Kind == "mo" && (textContent(n) == "∥" || textContent(n) == "∦") {
					v, _ := n.Attributes.GetExplicit("braketbar")
					got = append(got, v == true)
				}
				return true
			})
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("bar ownership %v, want %v", got, c.want)
			}
		})
	}
}

func TestBraketRegisteredMacrosWin(t *testing.T) {
	for _, name := range []string{"|", "Braket", "Set", "set"} {
		t.Run(name, func(t *testing.T) {
			state := newParseState()
			state.macros[name] = macroDefinition{body: "q"}
			p := &parser{source: "\\" + name, state: state}
			nodes, _, err := p.parseRow(0, false)
			if err != nil || len(nodes) != 1 || nodes[0].Kind != "mi" || textContent(nodes[0]) != "q" {
				t.Fatalf("registered macro lost: %v %v", nodes, err)
			}
		})
	}
}

func TestBraketOutsideBarRawAttributes(t *testing.T) {
	root, err := NewCompiler().Compile(`\|`, false)
	if err != nil {
		t.Fatal(err)
	}
	var bar *mml.Node
	root.Walk(func(n *mml.Node) bool {
		if n.Kind == "mo" {
			bar = n
		}
		return true
	})
	if bar == nil || textContent(bar) != "∥" || len(bar.Attributes.ExplicitNames()) != 0 {
		t.Fatalf("ordinary Bar differs from retained compiled primary: %v", bar)
	}
	if cls, ok := bar.Property("texClass"); !ok || cls != mml.TeXClassOrd {
		t.Fatalf("own texClass=%v, present=%v", cls, ok)
	}
}

func TestBraketSingleBarCreationOrder(t *testing.T) {
	for _, name := range []string{"Braket", "Set", "set"} {
		t.Run(name, func(t *testing.T) {
			p := &parser{source: `\|+z`, state: newParseState()}
			nodes, err := p.braket(name)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, n := range p.state.operators {
				got = append(got, textContent(n))
			}
			want := []string{"⟨", "⟩", "∥"}
			if name == "Set" {
				want = []string{"{", "}", "∥"}
			}
			if name == "set" {
				want = []string{"∥", "{", "}"}
			}
			if !reflect.DeepEqual(got, want) || p.source[p.pos:] != "+z" || p.braketOwner != nil {
				t.Fatalf("order=%v cursor=%d owner=%v", got, p.pos, p.braketOwner)
			}
			if name != "set" && (len(nodes) != 3 || nodes[0].Children[1].Kind != "TeXAtom" || textContent(nodes[1]) != "∥") {
				t.Fatalf("single owner swallowed remaining Bar output: %v", nodes)
			}
		})
	}
}

func TestBraketOwnerRestoredAfterError(t *testing.T) {
	outer := &braketItem{stretchy: true, barmax: -1}
	p := &parser{source: `{\undefined}`, state: newParseState(), braketOwner: outer}
	if _, err := p.braket("Set"); err == nil || p.braketOwner != outer {
		t.Fatalf("error=%v owner restored=%v", err, p.braketOwner == outer)
	}
}

func TestBraketAutoOpenBarrier(t *testing.T) {
	owner := &braketItem{single: true, stretchy: true, barmax: -1}
	p := &parser{source: `x\|)`, state: newParseState(), braketOwner: owner}
	auto := &derivativeAutoOpen{}
	nodes, _, err := p.parseRowWithAutoOpen(0, false, false, auto)
	if err != nil || !auto.closed || p.pos != len(p.source) || p.braketOwner != owner {
		t.Fatalf("AutoOpen incomplete: error=%v closed=%v cursor=%d", err, auto.closed, p.pos)
	}
	for _, n := range nodes {
		if value, _ := n.Attributes.GetExplicit("braketbar"); value == true {
			t.Fatal("Braket owns AutoOpen's escaped pipe")
		}
	}
}
