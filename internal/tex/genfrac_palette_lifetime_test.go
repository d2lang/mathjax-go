// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"errors"
	"strings"
	"testing"
)

// These are Go regression guards derived from the observed child-parser and
// same-parser contracts, not additional captured primary renders. Large seeded
// counters make missing reset/carry/restore behavior observable without a
// passive observer. StarMacro's separate own charge (D121) is not asserted.
func TestGenfracPaletteLifetimeGuards(t *testing.T) {
	if maxMacros != 1000 {
		t.Fatal("review the seeded guards when the configured limit changes")
	}
	longChild := strings.Repeat(`\p`, maxMacros) + `\frac{\p}{x}`
	tests := []struct {
		name, route, source, errorID string
		seed, restoredCount          int
		checkCount, registerVec      bool
	}{
		{"mathfont-fresh", "mathfont", `\p\p`, "", 999, 999, true, false},
		{"mathfont-descendant-carry", "mathfont", longChild, "", 999, 999, true, false},
		{"mathfont-error-restore", "mathfont", `\p\notDefined`, "UndefinedControlSequence", 999, 999, true, false},
		{"vector-fresh", "vector", `\p\p`, "", 999, 999, true, false},
		{"vector-descendant-carry", "vector", longChild, "", 999, 999, true, false},
		{"vector-error-restore", "vector", `\p\notDefined`, "UndefinedControlSequence", 999, 999, true, false},
		{"internalmath-descendant-carry", "internalmath", longChild, "", 999, 999, true, false},
		{"star-descendant-carry", "row", `\va{\p\p}`, "", 999, 0, false, false},
		{"star-continuation-overflow", "row", `\va{x}`, "MaxMacroSub1", 1000, 0, false, true},
		{"sqrtindex-fresh", "row", `\sqrt[\p\p]{x}`, "", 999, 999, true, false},
		{"sqrtindex-error-restore", "row", `\sqrt[\p\notDefined]{x}`, "UndefinedControlSequence", 999, 999, true, false},
		{"explicit-root-fresh", "row", `\root\p\p\of{x}`, "", 999, 999, true, false},
		{"explicit-root-error-restore", "row", `\root\p\notDefined\of{x}`, "UndefinedControlSequence", 999, 999, true, false},
		{"sameparser-group-style-overflow", "row", `{\bf\p}\scriptstyle\p`, "MaxMacroSub1", 999, 1001, true, false},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = c.seed
			state.macros["p"] = macroDefinition{body: "x"}
			if c.registerVec {
				state.macros["vec"] = macroDefinition{body: "x", arguments: 1}
			}
			p := &parser{source: "outer", pos: 2, state: state, genfracPalette: true}
			var err error
			switch c.route {
			case "mathfont":
				_, err = p.parseMathFontString(c.source, "normal", false)
			case "vector":
				_, err = p.parseVectorString(c.source, false)
			case "internalmath":
				_, err = p.parseInternalMath(c.source)
			case "row":
				p.source, p.pos = c.source, 0
				var stop string
				_, stop, err = p.parseRow(0, false)
				if err == nil && stop != "" {
					t.Fatalf("unexpected stop %q", stop)
				}
			default:
				t.Fatal("unbound guard route")
			}
			if p.state != state {
				t.Fatal("child replaced shared parser state")
			}
			if c.route != "row" && (p.source != "outer" || p.pos != 2) {
				t.Fatal("child replaced caller source/cursor")
			}
			if c.checkCount && state.macroCount != c.restoredCount {
				t.Fatalf("caller count %d; want %d", state.macroCount, c.restoredCount)
			}
			if c.errorID == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var parseError *Error
			if !errors.As(err, &parseError) || parseError.ID != c.errorID {
				t.Fatalf("error %v; want %s", err, c.errorID)
			}
			// In the Star overflow case, a future correct own charge may fail
			// before registered vec is reached. Both preserve this contract.
		})
	}
}
