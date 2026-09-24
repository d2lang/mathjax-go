// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// These installed programs, cursor zero and 0->1 charges are retained original
// MathJax 3.2.2 PairedDelimiters observations for the three public join witnesses.
func TestPairedDelimiterSourceContinuation(t *testing.T) {
	tests := []struct{ name, tail, installed string }{
		{"ordinary", `{\alpha}y`, `(\alpha x)y`},
		{"star", `*{\alpha}y`, `\left(\alpha x\right)y`},
		{"size", `[\Big]{\alpha}y`, `\Bigl(\alpha x\Bigr)y`},
	}
	const prefix = `\DeclarePairedDelimiterX{\pair}[1]{(}{)}{#1x}\pair`
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			state.pairedDelimiters["pair"] = pairedDelimiter{open: "(", close: ")", body: "#1x", arguments: 1}
			p := &parser{source: prefix + c.tail, pos: len(prefix), state: state}
			nodes, handled, err := p.mathtoolsCommand("pair")
			if err != nil || !handled {
				t.Fatalf("paired handler: handled=%v, error=%v", handled, err)
			}
			if len(nodes) != 0 || p.state != state {
				t.Fatal("paired handler returned parsed nodes or replaced caller state")
			}
			if p.source != c.installed || p.pos != 0 || p.state.macroCount != 1 {
				t.Fatalf("caller source=%q cursor=%d count=%d; want %q, 0, 1", p.source, p.pos, p.state.macroCount, c.installed)
			}
		})
	}
}
