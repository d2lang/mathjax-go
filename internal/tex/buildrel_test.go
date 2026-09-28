// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// ParseUpTo and ParseArg each own a fresh TexParser macro counter. The
// configuration remains shared and the caller's counter survives either child.
func TestBuildRelIndependentChildCounters(t *testing.T) {
	for _, source := range []string{`\probe\over\probe`, `\probe\over\undefined`, `\undefined\over\probe`} {
		t.Run(source, func(t *testing.T) {
			state := newParseState()
			state.macroCount = maxMacros
			state.macros["probe"] = macroDefinition{body: "x"}
			p := &parser{source: source, state: state}
			_, err := p.buildRelation("buildrel")
			if source == `\probe\over\probe` && err != nil {
				t.Fatalf("independent child Macro failed: %v", err)
			}
			if source != `\probe\over\probe` && err == nil {
				t.Fatal("missing child parser error")
			}
			if p.state != state || state.macroCount != maxMacros {
				t.Fatalf("caller state/count not restored: %d", state.macroCount)
			}
		})
	}
}
