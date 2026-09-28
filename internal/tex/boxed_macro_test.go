// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// AMS boxed is a Macro in the caller, followed by FBox/internalMath's fresh
// child parser. The child shares definitions but not the caller's counter.
func TestBoxedMacroCounterBoundary(t *testing.T) {
	for _, c := range []struct {
		source        string
		before, after int
		errorID       string
	}{
		{`\boxed{\probe}`, maxMacros - 1, maxMacros, ""},
		{`\boxed{\undefined}`, maxMacros - 1, maxMacros, "UndefinedControlSequence"},
		{`\boxed{x}`, maxMacros, maxMacros + 1, "MaxMacroSub1"},
		{`\boxed`, maxMacros, maxMacros, "MissingArgFor"},
	} {
		t.Run(c.source, func(t *testing.T) {
			state := newParseState()
			state.macros["probe"] = macroDefinition{body: "x"}
			state.macroCount = c.before
			p := &parser{source: c.source, state: state}
			_, _, err := p.parseRow(0, false)
			if c.errorID == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if e, ok := err.(*Error); !ok || e.ID != c.errorID {
				t.Fatalf("error = %v, want %s", err, c.errorID)
			}
			if p.state != state || state.macroCount != c.after {
				t.Fatalf("caller state/count = %d, want %d", state.macroCount, c.after)
			}
		})
	}
}
