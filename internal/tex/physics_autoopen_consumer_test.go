// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"errors"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// These are the actual parser inputs and source/node contracts retained in
// d099-d089-consumer-capture (freeze 38b36eb7). They exercise real registered
// macro expansion without renderer calls or whole-output expected snapshots.
// Complete operatorname rendering still has separate D095/D106 discrepancies.
func TestDerivativeAutoOpenConsumerMacroClose(t *testing.T) {
	const input = `\operatorname{a\dv{f}{x}(g\tailclose b}+Z`
	state := newParseState()
	state.macros["tailclose"] = macroDefinition{body: ")"}
	p := &parser{source: input, state: state, display: true}
	nodes, stop, err := p.parseRow(0, false)
	if err != nil {
		t.Fatal(err)
	}
	if stop != "" || p.source != input || p.pos != len(input) || p.state != state || state.macroCount != 1 {
		t.Fatalf("outer source/cursor/state changed: %q %d %q %d", p.source, p.pos, stop, state.macroCount)
	}
	if len(nodes) != 3 || nodes[0].Kind != "TeXAtom" || textContent(nodes[1]) != "+" || textContent(nodes[2]) != "Z" {
		t.Fatal("operator or outer continuation duplicated/lost")
	}
	if len(nodes[0].Children) != 1 || !nodes[0].Children[0].Flags.Inferred {
		t.Fatal("operator owner lost its inferred row")
	}
	body := nodes[0].Children[0]
	// The tail macro changes the child source to ")b" and leaves cursor 1.
	// Reading the old source or transferring its cursor before draining would
	// repeat, omit, or flatten these four distinct children.
	if len(body.Children) != 4 || body.Children[0].Kind != "mi" || textContent(body.Children[0]) != "a" || body.Children[1].Kind != "mfrac" || body.Children[2].Kind != "mrow" || textContent(body.Children[2]) != "(g)" || body.Children[3].Kind != "mi" || textContent(body.Children[3]) != "b" {
		t.Fatal("paired source/cursor must transfer after the fraction's action drains")
	}
	assertAutoOpenConsumerOwnership(t, nodes)
}

func TestDerivativeAutoOpenConsumerErrorOwnership(t *testing.T) {
	for _, c := range []struct {
		name, input, macro, expansion string
		cursor                        int
		color                         bool
	}{
		{"outer-caller-error", `Q+\operatorname{a\dv{f}{x}(g\tailbad)b}+Z`, "tailbad", `\undefinedTail`, 39, false},
		{"operator-ignore-error", `\operatorname{a\pdv[2]{f}{x}(\tailmark\undefinedTail)b}+Z`, "tailmark", `\definecolor{tailcolor}{RGB}{255,0,0}g`, 55, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			state.macros[c.macro] = macroDefinition{body: c.expansion}
			p := &parser{source: c.input, state: state, display: true}
			_, _, err := p.parseRow(0, false)
			var pe *Error
			if !errors.As(err, &pe) || pe.ID != "UndefinedControlSequence" || pe.Message != `Undefined control sequence \undefinedTail` {
				t.Fatalf("tail error suppressed or replaced: %v", err)
			}
			if p.source != c.input || p.pos != c.cursor || p.source[p.pos:] != "+Z" || p.state != state || state.macroCount != 1 {
				t.Fatalf("child error changed outer source/cursor/shared state: %q %d %d", p.source, p.pos, state.macroCount)
			}
			if c.color {
				color, err := state.colorModel.GetColor("named", "tailcolor")
				if err != nil || color != "#ff0000" {
					t.Fatal("ignored tail discarded its pre-error shared mutation", color, err)
				}
			}
		})
	}
}

func TestDerivativeAutoOpenConsumerPrimeRecipient(t *testing.T) {
	state := newParseState()
	state.macros["tailclose"] = macroDefinition{body: ")"}
	p := &parser{source: `x'^\dv{f}{x}(g\tailclose+Z`, state: state, display: true}
	nodes, stop, err := p.parseRow(0, false)
	if err != nil {
		t.Fatal(err)
	}
	if stop != "" || p.source != ")+Z" || p.pos != 3 || p.state != state || state.macroCount != 1 {
		t.Fatalf("macro-reset continuation changed: %q %d %q %d", p.source, p.pos, stop, state.macroCount)
	}
	if len(nodes) != 4 || nodes[0].Kind != "msup" || nodes[1].Kind != "mrow" || textContent(nodes[1]) != "(g)" || textContent(nodes[2]) != "+" || textContent(nodes[3]) != "Z" {
		t.Fatal("AutoOpen replacement must follow the prime script recipient")
	}
	if len(nodes[0].Children) != 2 || textContent(nodes[0].Children[0]) != "x" {
		t.Fatal("prime script base changed")
	}
	sup := nodes[0].Children[1]
	if sup.Kind != "mrow" || len(sup.Children) != 2 || textContent(sup.Children[0]) != "′" || sup.Children[1].Kind != "mfrac" {
		t.Fatal("prime/fraction recipient consumed the tail")
	}
	assertAutoOpenConsumerOwnership(t, nodes)
}

func assertAutoOpenConsumerOwnership(t *testing.T, roots []*mml.Node) {
	t.Helper()
	seen := map[*mml.Node]bool{}
	var visit func(*mml.Node, *mml.Node)
	visit = func(n, parent *mml.Node) {
		if seen[n] || n.Parent != parent {
			t.Fatal("duplicated or reparented delivered node")
		}
		seen[n] = true
		for _, child := range n.Children {
			visit(child, n)
		}
	}
	for _, root := range roots {
		visit(root, nil)
	}
}
