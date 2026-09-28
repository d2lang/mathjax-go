// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

func TestLeftRightRestoresCallerOnCloseAndError(t *testing.T) {
	for _, c := range []struct {
		name, source, errorID string
	}{
		{"right", `(\color{red}x\right)tail`, ""},
		{"middle", `(\color{red}\bf x\middle|y\over z\right)tail`, ""},
		{"nested", `(\left[\color{blue}x\middle|y\right]\middle|z\right)tail`, ""},
		{"position-error", `(\color{red}\raise1em\middle|y\right)`, "MissingBoxFor"},
		{"delimiter-error", `(\color{red}x\middle\undefined`, "MissingOrUnrecognizedDelim"},
		{"missing-right", `(\color{red}x\middle|y`, "ExtraLeftMissingRight"},
	} {
		t.Run(c.name, func(t *testing.T) {
			outer := &rowDelimiterItem{delimiter: "]", color: "blue"}
			p := &parser{source: c.source, state: newParseState(), rowDelimiter: outer,
				activeColor: "green", activeFont: "sans-serif", matrixClose: true, inRoot: true}
			_, err := p.leftRight("left")
			if c.errorID == "" {
				if err != nil || p.source[p.pos:] != "tail" {
					t.Fatalf("left/right consumed caller suffix: remaining %q, error %v", p.source[p.pos:], err)
				}
			} else if got, ok := err.(*Error); !ok || got.ID != c.errorID {
				t.Fatalf("error %v; want %s", err, c.errorID)
			}
			if p.rowDelimiter != outer || p.activeColor != "green" || p.activeFont != "sans-serif" || !p.matrixClose || !p.inRoot {
				t.Fatal("left/right did not restore its caller's delimiter, lexical environment, or close owner")
			}
		})
	}
}
