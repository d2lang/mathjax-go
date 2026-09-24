// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"errors"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// These seven registered inputs and their original-node delivery contracts
// were observed in MathJax 3.2.2's D2 profile. See the D095 finite capture and
// null-observer continuation, contract-traces.json: empty PushAll makes no
// stack push; single and multiple results forward the original nodes in order.
// These guards inspect actual Go pointers without adding a parser observer.
func TestDerivativeRegisteredResultDelivery(t *testing.T) {
	for _, c := range []struct {
		name, input, fraction, body string
		ignore                      bool
	}{
		{"frac-0-tail-dv", `\dv[2]{f}{x}(g)Z`, "frac", "", false},
		{"frac-0-tail-pdv", `\pdv[2]{f}{x}(g)Z`, "frac", "", true},
		{"frac-2-tail-dv", `\dv[2]{f}{x}(g)Z`, "frac", "xy", false},
		{"frac-2-tail-pdv", `\pdv[2]{f}{x}(g)Z`, "frac", "xy", true},
		{"frac-single-tail", `\dv[2]{f}{x}(g)Z`, "frac", "x", false},
		{"flatfrac-0-tail-dv", `\dv*[2]{f}{x}(g)Z`, "flatfrac", "", false},
		{"flatfrac-2-tail-dv", `\dv*[2]{f}{x}(g)Z`, "flatfrac", "xy", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			state.macros[c.fraction] = macroDefinition{body: c.body, arguments: 2}
			p := &parser{source: c.input, pos: 1, state: state, display: true}
			result, err := p.commandEvent(p.readControlSequence())
			if err != nil {
				t.Fatal(err)
			}
			if p.state != state || state.macroCount != 0 || p.source != c.input || p.source[p.pos:] != "(g)Z" {
				t.Fatal("fraction child changed the original parser before delivery")
			}
			if result.afterNode == nil || result.afterNode.ignore != c.ignore || len(result.nodes) != len(c.body) {
				t.Fatal("registered result or deferred action cardinality changed")
			}
			original := append([]*mml.Node(nil), result.nodes...)
			var returnedRow *mml.Node
			if len(original) == 2 {
				// The returned inferred row retains its child references during
				// PushAll. No transient detach/clear rule is imposed.
				returnedRow = original[0].Parent
				if returnedRow == nil || !returnedRow.Flags.Inferred || len(returnedRow.Children) != 2 {
					t.Fatal("missing original parsed inferred row")
				}
			}
			for i, n := range original {
				if n.Kind != "mi" || textContent(n) != c.body[i:i+1] {
					t.Fatal("fraction override was reconstructed or reordered")
				}
				if returnedRow != nil && (returnedRow.Children[i] != n || n.Parent != returnedRow) {
					t.Fatal("PushAll lost the parsed child identity")
				}
			}
			delivered := append([]*mml.Node(nil), result.nodes...)
			tail, err := result.afterNode.complete(p)
			if err != nil {
				t.Fatal(err)
			}
			if p.source[p.pos:] != "Z" || !result.afterNode.closed || result.afterNode.openCount != -1 || (len(tail) == 0) != c.ignore {
				t.Fatal("AutoOpen did not follow result delivery on the original parser")
			}
			if !c.ignore && (len(tail) != 1 || textContent(tail[0]) != "(g)") {
				t.Fatal("wrong following AutoOpen content")
			}
			delivered = append(delivered, tail...)
			remainder, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" || len(remainder) != 1 || textContent(remainder[0]) != "Z" {
				t.Fatal("wrong outer-parser remainder", err, stop)
			}
			delivered = append(delivered, remainder...)
			recipient := forcedRow(delivered, true)
			for i, n := range original {
				if recipient.Children[i] != n || n.Parent != recipient {
					t.Fatal("final attachment cloned or lost the forwarded node")
				}
				if returnedRow != nil && returnedRow.Children[i] != n {
					t.Fatal("forwarding cleared the original inferred-row children")
				}
			}
		})
	}
}

// Passive primary 15a8de27d341757045d24e843752e67ffafae236b4887c61b70030dc2a76d82b:
// zero children are delivered first; the pending script rejects auto open at
// outer cursor 16, before reading g. In particular, an empty fraction result
// cannot silently replace the superscript with the following parenthesis.
func TestDerivativeEmptyResultPendingPrime(t *testing.T) {
	state := newParseState()
	state.macros["frac"] = macroDefinition{body: "", arguments: 2}
	input := `x'^\dv[2]{f}{x}(g)Z`
	p := &parser{source: input, state: state, display: true}
	_, _, err := p.parseRow(0, false)
	var parseError *Error
	if !errors.As(err, &parseError) || parseError.ID != "MissingOpenForSup" || parseError.Error() != "Missing open brace for superscript" {
		t.Fatalf("pending recipient error = %v", err)
	}
	if p.state != state || state.macroCount != 0 || p.source != input || p.pos != 16 || p.source[p.pos:] != "g)Z" {
		t.Fatal("auto-open rejection consumed content or changed the original parser")
	}
}

// The exact 600-name acyclic registration runs twice in distinct primary
// fraction argument parsers (passive d6180b3c9233a638f2516dd85d56ec7fd0960375ae4ff1e751ea322e8069de6e).
// Grouping both uses in one argument instead raises MaxMacroSub1 on its 1001st
// substitution (passive 97a030552f074b3f2b07ff7c3b7be8fe3fd52303de884331a4ff5db5d63dfa53).
// Both leave the original Derivative parser at count 999. This distinguishes
// child-parser budgets from a reset for every Go helper or brace group.
func TestDerivativeIndependentOrderBudgets(t *testing.T) {
	if maxMacros != 1000 {
		t.Fatal("review the observed budget contract when the limit changes")
	}
	for _, c := range []struct {
		name, input, errorID string
	}{
		{"independent-arguments", `\dv[\dchaina]{f}{x}Z`, ""},
		{"same-parser-groups", `\dv[{\dchaina}{\dchaina}]{f}{x}Z`, "MaxMacroSub1"},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			for i := 0; i < 600; i++ {
				body := "n"
				if i < 599 {
					body = "\\" + derivativeChainName(i+1)
				}
				state.macros[derivativeChainName(i)] = macroDefinition{body: body}
			}
			state.macroCount = 999
			p := &parser{source: c.input, pos: 1, state: state, display: true}
			result, err := p.commandEvent(p.readControlSequence())
			if p.state != state || state.macroCount != 999 || p.source != c.input || p.source[p.pos:] != "Z" {
				t.Fatal("child parsing did not restore the caller's state/count/cursor")
			}
			if c.errorID != "" {
				var parseError *Error
				if !errors.As(err, &parseError) || parseError.ID != c.errorID || result.afterNode != nil || len(result.nodes) != 0 {
					t.Fatal("same-parser chain error or pre-delivery state changed", err)
				}
				return
			}
			if err != nil || len(result.nodes) != 1 || result.nodes[0].Kind != "mfrac" || result.afterNode == nil {
				t.Fatal("independent argument chains failed", err)
			}
			var orders []*mml.Node
			result.nodes[0].Walk(func(n *mml.Node) bool {
				if n.Kind == "mi" && textContent(n) == "n" {
					orders = append(orders, n)
				}
				return true
			})
			if len(orders) != 2 || orders[0] == orders[1] {
				t.Fatal("the two order occurrences did not produce distinct nodes")
			}
		})
	}
}

// Spreadsheet-style alphabetic suffixes reproduce the saved registration:
// dchaina, ..., dchainz, dchainaa, ..., dchainwb, whose body is n.
func derivativeChainName(index int) string {
	name := ""
	for index++; index > 0; index /= 26 {
		index--
		name = string(rune('a'+index%26)) + name
	}
	return "dchain" + name
}
