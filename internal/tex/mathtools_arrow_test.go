// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// ArrowBetweenLines reparses its expansion in a new TexParser, so the child
// has its own macro budget and restores the caller's budget even on error.
func TestArrowBetweenLinesChildCounter(t *testing.T) {
	for _, source := range []string{`\ArrowBetweenLines[\probe]`, `\ArrowBetweenLines[\undefined]`} {
		t.Run(source, func(t *testing.T) {
			state := newParseState()
			state.macroCount = maxMacros
			state.macros["probe"] = macroDefinition{body: "x"}
			p := &parser{state: state}
			rows := 0
			err := p.parseEquationRow([]string{source}, true, func(*mml.Node) error {
				rows++
				return nil
			})
			if source == `\ArrowBetweenLines[\probe]` && (err != nil || rows != 1) {
				t.Fatalf("independent child did not emit its row: rows=%d, err=%v", rows, err)
			}
			if source == `\ArrowBetweenLines[\undefined]` && (err == nil || rows != 0) {
				t.Fatalf("child error emitted a row: rows=%d, err=%v", rows, err)
			}
			if p.state != state || state.macroCount != maxMacros {
				t.Fatalf("caller count not restored: %d", state.macroCount)
			}
		})
	}
}
