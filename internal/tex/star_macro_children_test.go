// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// These are Go guards derived from the two observed genuine-child boundaries,
// not additional primary renders. A real Macro at the maximum caller budget
// must succeed in the child and restore that caller budget on return.
func TestStarMacroChildCounters(t *testing.T) {
	for _, route := range []string{"argument", "vector"} {
		t.Run(route, func(t *testing.T) {
			state := newParseState()
			state.macroCount = maxMacros
			state.macros["probe"] = macroDefinition{body: "x"}
			p := &parser{source: `{\probe}`, state: state, starMacroChildren: true}
			var err error
			if route == "argument" {
				_, err = p.parseArgument("vec")
			} else {
				_, err = p.parseVectorString(`\probe`, false)
			}
			if err != nil {
				t.Fatalf("fresh child Macro failed: %v", err)
			}
			if p.state != state || state.macroCount != maxMacros {
				t.Fatalf("caller state/count not restored: %d", state.macroCount)
			}
		})
	}
}
